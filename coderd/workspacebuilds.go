package coderd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"cdr.dev/slog"

	"github.com/coder/coder/v2/coderd/audit"
	"github.com/coder/coder/v2/coderd/database"
	"github.com/coder/coder/v2/coderd/database/db2sdk"
	"github.com/coder/coder/v2/coderd/database/dbauthz"
	"github.com/coder/coder/v2/coderd/database/dbtime"
	"github.com/coder/coder/v2/coderd/database/provisionerjobs"
	"github.com/coder/coder/v2/coderd/httpapi"
	"github.com/coder/coder/v2/coderd/httpapi/httperror"
	"github.com/coder/coder/v2/coderd/httpmw"
	"github.com/coder/coder/v2/coderd/notifications"
	"github.com/coder/coder/v2/coderd/provisionerdserver"
	"github.com/coder/coder/v2/coderd/rbac"
	"github.com/coder/coder/v2/coderd/rbac/policy"
	"github.com/coder/coder/v2/coderd/wsbuilder"
	"github.com/coder/coder/v2/coderd/wspubsub"
	"github.com/coder/coder/v2/codersdk"
)

// @Summary Get workspace build
// @ID get-workspace-build
// @Security CoderSessionToken
// @Produce json
// @Tags Builds
// @Param workspacebuild path string true "Workspace build ID"
// @Success 200 {object} codersdk.WorkspaceBuild
// @Router /workspacebuilds/{workspacebuild} [get]
func (api *API) workspaceBuild(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceBuild := httpmw.WorkspaceBuildParam(r)
	workspace := httpmw.WorkspaceParam(r)

	data, err := api.workspaceBuildsData(ctx, []database.WorkspaceBuild{workspaceBuild})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error getting workspace build data.",
			Detail:  err.Error(),
		})
		return
	}

	// Ensure we have the job and template version for the workspace build.
	// Otherwise we risk a panic in the api.convertWorkspaceBuild call below.
	if len(data.jobs) == 0 {
		httpapi.Write(ctx, rw, http.StatusNotFound, codersdk.Response{
			Message: "Internal error getting workspace build data.",
			Detail:  "No job found for workspace build.",
		})
		return
	}

	if len(data.templateVersions) == 0 {
		httpapi.Write(ctx, rw, http.StatusNotFound, codersdk.Response{
			Message: "Internal error getting workspace build data.",
			Detail:  "No template version found for workspace build.",
		})
		return
	}

	apiBuild, err := api.convertWorkspaceBuild(
		workspaceBuild,
		workspace,
		data.jobs[0],
		data.resources,
		data.metadata,
		data.agents,
		data.apps,
		data.appStatuses,
		data.scripts,
		data.logSources,
		data.templateVersions[0],
		nil,
	)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error converting workspace build.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, apiBuild)
}

// @Summary Get workspace builds by workspace ID
// @ID get-workspace-builds-by-workspace-id
// @Security CoderSessionToken
// @Produce json
// @Tags Builds
// @Param workspace path string true "Workspace ID" format(uuid)
// @Param after_id query string false "After ID" format(uuid)
// @Param limit query int false "Page limit"
// @Param offset query int false "Page offset"
// @Param since query string false "Since timestamp" format(date-time)
// @Success 200 {array} codersdk.WorkspaceBuild
// @Router /workspaces/{workspace}/builds [get]
func (api *API) workspaceBuilds(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspace := httpmw.WorkspaceParam(r)

	paginationParams, ok := ParsePagination(rw, r)
	if !ok {
		return
	}

	var since time.Time

	sinceParam := r.URL.Query().Get("since")
	if sinceParam != "" {
		var err error
		since, err = time.Parse(time.RFC3339, sinceParam)
		if err != nil {
			httpapi.Write(r.Context(), rw, http.StatusBadRequest, codersdk.Response{
				Message: "bad `since` format, must be RFC3339",
				Detail:  err.Error(),
			})
			return
		}
	}

	var workspaceBuilds []database.WorkspaceBuild
	// Ensure all db calls happen in the same tx
	err := api.Database.InTx(func(store database.Store) error {
		var err error
		if paginationParams.AfterID != uuid.Nil {
			// See if the record exists first. If the record does not exist, the pagination
			// query will not work.
			_, err := store.GetWorkspaceBuildByID(ctx, paginationParams.AfterID)
			if err != nil && xerrors.Is(err, sql.ErrNoRows) {
				httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
					Message: fmt.Sprintf("Record at \"after_id\" (%q) does not exist.", paginationParams.AfterID.String()),
				})
				return err
			} else if err != nil {
				httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
					Message: "Internal error fetching workspace build at \"after_id\".",
					Detail:  err.Error(),
				})
				return err
			}
		}

		req := database.GetWorkspaceBuildsByWorkspaceIDParams{
			WorkspaceID: workspace.ID,
			AfterID:     paginationParams.AfterID,
			// #nosec G115 - Pagination offsets are small and fit in int32
			OffsetOpt: int32(paginationParams.Offset),
			// #nosec G115 - Pagination limits are small and fit in int32
			LimitOpt: int32(paginationParams.Limit),
			Since:    dbtime.Time(since),
		}
		workspaceBuilds, err = store.GetWorkspaceBuildsByWorkspaceID(ctx, req)
		if xerrors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		if err != nil {
			httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
				Message: "Internal error fetching workspace build.",
				Detail:  err.Error(),
			})
			return err
		}

		return nil
	}, nil)
	if err != nil {
		return
	}

	data, err := api.workspaceBuildsData(ctx, workspaceBuilds)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error getting workspace build data.",
			Detail:  err.Error(),
		})
		return
	}

	apiBuilds, err := api.convertWorkspaceBuilds(
		workspaceBuilds,
		[]database.Workspace{workspace},
		data.jobs,
		data.resources,
		data.metadata,
		data.agents,
		data.apps,
		data.appStatuses,
		data.scripts,
		data.logSources,
		data.templateVersions,
		data.provisionerDaemons,
	)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error converting workspace build.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, apiBuilds)
}

// @Summary Get workspace build by user, workspace name, and build number
// @ID get-workspace-build-by-user-workspace-name-and-build-number
// @Security CoderSessionToken
// @Produce json
// @Tags Builds
// @Param user path string true "User ID, name, or me"
// @Param workspacename path string true "Workspace name"
// @Param buildnumber path string true "Build number" format(number)
// @Success 200 {object} codersdk.WorkspaceBuild
// @Router /users/{user}/workspace/{workspacename}/builds/{buildnumber} [get]
func (api *API) workspaceBuildByBuildNumber(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	mems := httpmw.OrganizationMembersParam(r)
	workspaceName := chi.URLParam(r, "workspacename")
	buildNumber, err := strconv.ParseInt(chi.URLParam(r, "buildnumber"), 10, 32)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Failed to parse build number as integer.",
			Detail:  err.Error(),
		})
		return
	}

	workspace, err := api.Database.GetWorkspaceByOwnerIDAndName(ctx, database.GetWorkspaceByOwnerIDAndNameParams{
		OwnerID: mems.UserID(),
		Name:    workspaceName,
	})
	if httpapi.Is404Error(err) {
		httpapi.ResourceNotFound(rw)
		return
	}
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error fetching workspace by name.",
			Detail:  err.Error(),
		})
		return
	}

	workspaceBuild, err := api.Database.GetWorkspaceBuildByWorkspaceIDAndBuildNumber(ctx, database.GetWorkspaceBuildByWorkspaceIDAndBuildNumberParams{
		WorkspaceID: workspace.ID,
		BuildNumber: int32(buildNumber),
	})
	if httpapi.Is404Error(err) {
		httpapi.Write(ctx, rw, http.StatusNotFound, codersdk.Response{
			Message: fmt.Sprintf("Workspace %q Build %d does not exist.", workspaceName, buildNumber),
		})
		return
	}
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error fetching workspace build.",
			Detail:  err.Error(),
		})
		return
	}

	data, err := api.workspaceBuildsData(ctx, []database.WorkspaceBuild{workspaceBuild})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error getting workspace build data.",
			Detail:  err.Error(),
		})
		return
	}

	apiBuild, err := api.convertWorkspaceBuild(
		workspaceBuild,
		workspace,
		data.jobs[0],
		data.resources,
		data.metadata,
		data.agents,
		data.apps,
		data.appStatuses,
		data.scripts,
		data.logSources,
		data.templateVersions[0],
		data.provisionerDaemons,
	)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error converting workspace build.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, apiBuild)
}

// Azure supports instance identity verification:
// https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service?tabs=linux#tabgroup_14
//
// @Summary Create workspace build
// @ID create-workspace-build
// @Security CoderSessionToken
// @Accept json
// @Produce json
// @Tags Builds
// @Param workspace path string true "Workspace ID" format(uuid)
// @Param request body codersdk.CreateWorkspaceBuildRequest true "Create workspace build request"
// @Success 200 {object} codersdk.WorkspaceBuild
// @Router /workspaces/{workspace}/builds [post]
func (api *API) postWorkspaceBuilds(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	apiKey := httpmw.APIKey(r)

	workspace := httpmw.WorkspaceParam(r)
	var createBuild codersdk.CreateWorkspaceBuildRequest
	if !httpapi.Read(ctx, rw, r, &createBuild) {
		return
	}

	transition := database.WorkspaceTransition(createBuild.Transition)
	builder := wsbuilder.New(workspace, transition, *api.BuildUsageChecker.Load()).
		Initiator(apiKey.UserID).
		RichParameterValues(createBuild.RichParameterValues).
		LogLevel(string(createBuild.LogLevel)).
		DeploymentValues(api.Options.DeploymentValues).
		Experiments(api.Experiments).
		TemplateVersionPresetID(createBuild.TemplateVersionPresetID)

	if transition == database.WorkspaceTransitionStart && createBuild.Reason != "" {
		builder = builder.Reason(database.BuildReason(createBuild.Reason))
	}

	var (
		previousWorkspaceBuild database.WorkspaceBuild
		workspaceBuild         *database.WorkspaceBuild
		provisionerJob         *database.ProvisionerJob
		provisionerDaemons     []database.GetEligibleProvisionerDaemonsByProvisionerJobIDsRow
	)

	err := api.Database.InTx(func(tx database.Store) error {
		var err error

		previousWorkspaceBuild, err = tx.GetLatestWorkspaceBuildByWorkspaceID(ctx, workspace.ID)
		if err != nil && !xerrors.Is(err, sql.ErrNoRows) {
			api.Logger.Error(ctx, "failed fetching previous workspace build", slog.F("workspace_id", workspace.ID), slog.Error(err))
			httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
				Message: "Internal error fetching previous workspace build",
				Detail:  err.Error(),
			})
			return nil
		}

		if createBuild.TemplateVersionID != uuid.Nil {
			builder = builder.VersionID(createBuild.TemplateVersionID)
		}

		if createBuild.Orphan {
			if createBuild.Transition != codersdk.WorkspaceTransitionDelete {
				httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
					Message: "Orphan is only permitted when deleting a workspace.",
				})
				return nil
			}
			if len(createBuild.ProvisionerState) > 0 {
				httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
					Message: "ProvisionerState cannot be set alongside Orphan since state intent is unclear.",
				})
				return nil
			}
			builder = builder.Orphan()
		}
		if len(createBuild.ProvisionerState) > 0 {
			builder = builder.State(createBuild.ProvisionerState)
		}

		workspaceBuild, provisionerJob, provisionerDaemons, err = builder.Build(
			ctx,
			tx,
			api.FileCache,
			func(action policy.Action, object rbac.Objecter) bool {
				if auth := api.Authorize(r, action, object); auth {
					return true
				}
				// Special handling for prebuilt workspace deletion
				if action == policy.ActionDelete {
					if workspaceObj, ok := object.(database.PrebuiltWorkspaceResource); ok && workspaceObj.IsPrebuild() {
						return api.Authorize(r, action, workspaceObj.AsPrebuild())
					}
				}
				return false
			},
			audit.WorkspaceBuildBaggageFromRequest(r),
		)
		return err
	}, nil)
	if err != nil {
		httperror.WriteWorkspaceBuildError(ctx, rw, err)
		return
	}

	var queuePos database.GetProvisionerJobsByIDsWithQueuePositionRow
	if provisionerJob != nil {
		queuePos.ProvisionerJob = *provisionerJob
		queuePos.QueuePosition = 0
		if err := provisionerjobs.PostJob(api.Pubsub, *provisionerJob); err != nil {
			// Client probably doesn't care about this error, so just log it.
			api.Logger.Error(ctx, "failed to post provisioner job to pubsub", slog.Error(err))
		}

		// We may need to complete the audit if wsbuilder determined that
		// no provisioner could handle an orphan-delete job and completed it.
		if createBuild.Orphan && createBuild.Transition == codersdk.WorkspaceTransitionDelete && provisionerJob.CompletedAt.Valid {
			api.Logger.Warn(ctx, "orphan delete handled by wsbuilder due to no eligible provisioners",
				slog.F("workspace_id", workspace.ID),
				slog.F("workspace_build_id", workspaceBuild.ID),
				slog.F("provisioner_job_id", provisionerJob.ID),
			)
			buildResourceInfo := audit.AdditionalFields{
				WorkspaceName:  workspace.Name,
				BuildNumber:    strconv.Itoa(int(workspaceBuild.BuildNumber)),
				BuildReason:    workspaceBuild.Reason,
				WorkspaceID:    workspace.ID,
				WorkspaceOwner: workspace.OwnerName,
			}
			briBytes, err := json.Marshal(buildResourceInfo)
			if err != nil {
				api.Logger.Error(ctx, "failed to marshal build resource info for audit", slog.Error(err))
			}
			auditor := api.Auditor.Load()
			bag := audit.BaggageFromContext(ctx)
			audit.BackgroundAudit(ctx, &audit.BackgroundAuditParams[database.WorkspaceBuild]{
				Audit:            *auditor,
				Log:              api.Logger,
				UserID:           provisionerJob.InitiatorID,
				OrganizationID:   workspace.OrganizationID,
				RequestID:        provisionerJob.ID,
				IP:               bag.IP,
				Action:           database.AuditActionDelete,
				Old:              previousWorkspaceBuild,
				New:              *workspaceBuild,
				Status:           http.StatusOK,
				AdditionalFields: briBytes,
			})
		}
	}

	apiBuild, err := api.convertWorkspaceBuild(
		*workspaceBuild,
		workspace,
		queuePos,
		[]database.WorkspaceResource{},
		[]database.WorkspaceResourceMetadatum{},
		[]database.WorkspaceAgent{},
		[]database.WorkspaceApp{},
		[]database.WorkspaceAppStatus{},
		[]database.WorkspaceAgentScript{},
		[]database.WorkspaceAgentLogSource{},
		database.TemplateVersion{},
		provisionerDaemons,
	)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error converting workspace build.",
			Detail:  err.Error(),
		})
		return
	}

	// If this workspace build has a different template version ID to the previous build
	// we can assume it has just been updated.
	if createBuild.TemplateVersionID != uuid.Nil && createBuild.TemplateVersionID != previousWorkspaceBuild.TemplateVersionID {
		// nolint:gocritic // Need system context to fetch admins
		admins, err := findTemplateAdmins(dbauthz.AsSystemRestricted(ctx), api.Database)
		if err != nil {
			api.Logger.Error(ctx, "find template admins", slog.Error(err))
		} else {
			for _, admin := range admins {
				// Don't send notifications to user which initiated the event.
				if admin.ID == apiKey.UserID {
					continue
				}

				api.notifyWorkspaceUpdated(ctx, apiKey.UserID, admin.ID, workspace, createBuild.RichParameterValues)
			}
		}
	}

	api.publishWorkspaceUpdate(ctx, workspace.OwnerID, wspubsub.WorkspaceEvent{
		Kind:        wspubsub.WorkspaceEventKindStateChange,
		WorkspaceID: workspace.ID,
	})

	httpapi.Write(ctx, rw, http.StatusCreated, apiBuild)
}

func (api *API) notifyWorkspaceUpdated(
	ctx context.Context,
	initiatorID uuid.UUID,
	receiverID uuid.UUID,
	workspace database.Workspace,
	parameters []codersdk.WorkspaceBuildParameter,
) {
	log := api.Logger.With(slog.F("workspace_id", workspace.ID))

	template, err := api.Database.GetTemplateByID(ctx, workspace.TemplateID)
	if err != nil {
		log.Warn(ctx, "failed to fetch template for workspace creation notification", slog.F("template_id", workspace.TemplateID), slog.Error(err))
		return
	}

	version, err := api.Database.GetTemplateVersionByID(ctx, template.ActiveVersionID)
	if err != nil {
		log.Warn(ctx, "failed to fetch template version for workspace creation notification", slog.F("template_id", workspace.TemplateID), slog.Error(err))
		return
	}

	initiator, err := api.Database.GetUserByID(ctx, initiatorID)
	if err != nil {
		log.Warn(ctx, "failed to fetch user for workspace update notification", slog.F("initiator_id", initiatorID), slog.Error(err))
		return
	}

	owner, err := api.Database.GetUserByID(ctx, workspace.OwnerID)
	if err != nil {
		log.Warn(ctx, "failed to fetch user for workspace update notification", slog.F("owner_id", workspace.OwnerID), slog.Error(err))
		return
	}

	buildParameters := make([]map[string]any, len(parameters))
	for idx, parameter := range parameters {
		buildParameters[idx] = map[string]any{
			"name":  parameter.Name,
			"value": parameter.Value,
		}
	}

	if _, err := api.NotificationsEnqueuer.EnqueueWithData(
		// nolint:gocritic // Need notifier actor to enqueue notifications
		dbauthz.AsNotifier(ctx),
		receiverID,
		notifications.TemplateWorkspaceManuallyUpdated,
		map[string]string{
			"organization":             template.OrganizationName,
			"initiator":                initiator.Name,
			"workspace":                workspace.Name,
			"template":                 template.Name,
			"version":                  version.Name,
			"workspace_owner_username": owner.Username,
		},
		map[string]any{
			"workspace":        map[string]any{"id": workspace.ID, "name": workspace.Name},
			"template":         map[string]any{"id": template.ID, "name": template.Name},
			"template_version": map[string]any{"id": version.ID, "name": version.Name},
			"owner":            map[string]any{"id": owner.ID, "name": owner.Name, "email": owner.Email},
			"parameters":       buildParameters,
		},
		"api-workspaces-updated",
		// Associate this notification with all the related entities
		workspace.ID, workspace.OwnerID, workspace.TemplateID, workspace.OrganizationID,
	); err != nil {
		log.Warn(ctx, "failed to notify of workspace update", slog.Error(err))
	}
}

// @Summary Cancel workspace build
// @ID cancel-workspace-build
// @Security CoderSessionToken
// @Produce json
// @Tags Builds
// @Param workspacebuild path string true "Workspace build ID"
// @Param expect_status query string false "Expected status of the job. If expect_status is supplied, the request will be rejected with 412 Precondition Failed if the job doesn't match the state when performing the cancellation." Enums(running, pending)
// @Success 200 {object} codersdk.Response
// @Router /workspacebuilds/{workspacebuild}/cancel [patch]
func (api *API) patchCancelWorkspaceBuild(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var expectStatus database.ProvisionerJobStatus
	expectStatusParam := r.URL.Query().Get("expect_status")
	if expectStatusParam != "" {
		if expectStatusParam != "running" && expectStatusParam != "pending" {
			httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
				Message: fmt.Sprintf("Invalid expect_status %q. Only 'running' or 'pending' are allowed.", expectStatusParam),
			})
			return
		}
		expectStatus = database.ProvisionerJobStatus(expectStatusParam)
	}

	workspaceBuild := httpmw.WorkspaceBuildParam(r)
	workspace, err := api.Database.GetWorkspaceByID(ctx, workspaceBuild.WorkspaceID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "No workspace exists for this job.",
		})
		return
	}

	code := http.StatusInternalServerError
	resp := codersdk.Response{
		Message: "Internal error canceling workspace build.",
	}
	err = api.Database.InTx(func(db database.Store) error {
		valid, err := verifyUserCanCancelWorkspaceBuilds(ctx, db, httpmw.APIKey(r).UserID, workspace.TemplateID, expectStatus)
		if err != nil {
			code = http.StatusInternalServerError
			resp.Message = "Internal error verifying permission to cancel workspace build."
			resp.Detail = err.Error()

			return xerrors.Errorf("verify user can cancel workspace builds: %w", err)
		}
		if !valid {
			code = http.StatusForbidden
			resp.Message = "User is not allowed to cancel workspace builds. Owner role is required."

			return xerrors.New("user is not allowed to cancel workspace builds")
		}

		job, err := db.GetProvisionerJobByIDForUpdate(ctx, workspaceBuild.JobID)
		if err != nil {
			code = http.StatusInternalServerError
			resp.Message = "Internal error fetching provisioner job."
			resp.Detail = err.Error()

			return xerrors.Errorf("get provisioner job: %w", err)
		}
		if job.CompletedAt.Valid {
			code = http.StatusBadRequest
			resp.Message = "Job has already completed!"

			return xerrors.New("job has already completed")
		}
		if job.CanceledAt.Valid {
			code = http.StatusBadRequest
			resp.Message = "Job has already been marked as canceled!"

			return xerrors.New("job has already been marked as canceled")
		}

		if expectStatus != "" && job.JobStatus != expectStatus {
			code = http.StatusPreconditionFailed
			resp.Message = "Job is not in the expected state."

			return xerrors.Errorf("job is not in the expected state: expected: %q, got %q", expectStatus, job.JobStatus)
		}

		err = db.UpdateProvisionerJobWithCancelByID(ctx, database.UpdateProvisionerJobWithCancelByIDParams{
			ID: job.ID,
			CanceledAt: sql.NullTime{
				Time:  dbtime.Now(),
				Valid: true,
			},
			CompletedAt: sql.NullTime{
				Time: dbtime.Now(),
				// If the job is running, don't mark it completed!
				Valid: !job.WorkerID.Valid,
			},
		})
		if err != nil {
			code = http.StatusInternalServerError
			resp.Message = "Internal error updating provisioner job."
			resp.Detail = err.Error()

			return xerrors.Errorf("update provisioner job: %w", err)
		}

		return nil
	}, nil)
	if err != nil {
		httpapi.Write(ctx, rw, code, resp)
		return
	}

	api.publishWorkspaceUpdate(ctx, workspace.OwnerID, wspubsub.WorkspaceEvent{
		Kind:        wspubsub.WorkspaceEventKindStateChange,
		WorkspaceID: workspace.ID,
	})

	httpapi.Write(ctx, rw, http.StatusOK, codersdk.Response{
		Message: "Job has been marked as canceled...",
	})
}

func verifyUserCanCancelWorkspaceBuilds(ctx context.Context, store database.Store, userID uuid.UUID, templateID uuid.UUID, jobStatus database.ProvisionerJobStatus) (bool, error) {
	// If the jobStatus is pending, we always allow cancellation regardless of
	// the template setting as it's non-destructive to Terraform resources.
	if jobStatus == database.ProvisionerJobStatusPending {
		return true, nil
	}

	template, err := store.GetTemplateByID(ctx, templateID)
	if err != nil {
		return false, xerrors.New("no template exists for this workspace")
	}

	if template.AllowUserCancelWorkspaceJobs {
		return true, nil // all users can cancel workspace builds
	}

	user, err := store.GetUserByID(ctx, userID)
	if err != nil {
		return false, xerrors.New("user does not exist")
	}
	return slices.Contains(user.RBACRoles, rbac.RoleOwner().String()), nil // only user with "owner" role can cancel workspace builds
}

// @Summary Get build parameters for workspace build
// @ID get-build-parameters-for-workspace-build
// @Security CoderSessionToken
// @Produce json
// @Tags Builds
// @Param workspacebuild path string true "Workspace build ID"
// @Success 200 {array} codersdk.WorkspaceBuildParameter
// @Router /workspacebuilds/{workspacebuild}/parameters [get]
func (api *API) workspaceBuildParameters(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceBuild := httpmw.WorkspaceBuildParam(r)

	parameters, err := api.Database.GetWorkspaceBuildParameters(ctx, workspaceBuild.ID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error fetching workspace build parameters.",
			Detail:  err.Error(),
		})
		return
	}
	apiParameters := db2sdk.WorkspaceBuildParameters(parameters)
	httpapi.Write(ctx, rw, http.StatusOK, apiParameters)
}

// @Summary Get workspace build logs
// @ID get-workspace-build-logs
// @Security CoderSessionToken
// @Produce json
// @Tags Builds
// @Param workspacebuild path string true "Workspace build ID"
// @Param before query int false "Before log id"
// @Param after query int false "After log id"
// @Param follow query bool false "Follow log stream"
// @Success 200 {array} codersdk.ProvisionerJobLog
// @Router /workspacebuilds/{workspacebuild}/logs [get]
func (api *API) workspaceBuildLogs(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceBuild := httpmw.WorkspaceBuildParam(r)

	job, err := api.Database.GetProvisionerJobByID(ctx, workspaceBuild.JobID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error fetching provisioner job.",
			Detail:  err.Error(),
		})
		return
	}
	api.provisionerJobLogs(rw, r, job)
}

// @Summary Get provisioner state for workspace build
// @ID get-provisioner-state-for-workspace-build
// @Security CoderSessionToken
// @Produce json
// @Tags Builds
// @Param workspacebuild path string true "Workspace build ID"
// @Success 200 {object} codersdk.WorkspaceBuild
// @Router /workspacebuilds/{workspacebuild}/state [get]
func (api *API) workspaceBuildState(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	workspaceBuild := httpmw.WorkspaceBuildParam(r)
	workspace, err := api.Database.GetWorkspaceByID(ctx, workspaceBuild.WorkspaceID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "No workspace exists for this job.",
		})
		return
	}
	template, err := api.Database.GetTemplateByID(ctx, workspace.TemplateID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Failed to get template",
			Detail:  err.Error(),
		})
		return
	}

	// You must have update permissions on the template to get the state.
	// This matches a push!
	if !api.Authorize(r, policy.ActionUpdate, template.RBACObject()) {
		httpapi.ResourceNotFound(rw)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(workspaceBuild.ProvisionerState)
}

// @Summary Get workspace build timings by ID
// @ID get-workspace-build-timings-by-id
// @Security CoderSessionToken
// @Produce json
// @Tags Builds
// @Param workspacebuild path string true "Workspace build ID" format(uuid)
// @Success 200 {object} codersdk.WorkspaceBuildTimings
// @Router /workspacebuilds/{workspacebuild}/timings [get]
func (api *API) workspaceBuildTimings(rw http.ResponseWriter, r *http.Request) {
	var (
		ctx   = r.Context()
		build = httpmw.WorkspaceBuildParam(r)
	)

	timings, err := api.buildTimings(ctx, build)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error fetching timings.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, timings)
}

type workspaceBuildsData struct {
	jobs               []database.GetProvisionerJobsByIDsWithQueuePositionRow
	templateVersions   []database.TemplateVersion
	resources          []database.WorkspaceResource
	metadata           []database.WorkspaceResourceMetadatum
	agents             []database.WorkspaceAgent
	apps               []database.WorkspaceApp
	appStatuses        []database.WorkspaceAppStatus
	scripts            []database.WorkspaceAgentScript
	logSources         []database.WorkspaceAgentLogSource
	provisionerDaemons []database.GetEligibleProvisionerDaemonsByProvisionerJobIDsRow
}

func (api *API) workspaceBuildsData(ctx context.Context, workspaceBuilds []database.WorkspaceBuild) (workspaceBuildsData, error) {
	jobIDs := make([]uuid.UUID, 0, len(workspaceBuilds))
	for _, build := range workspaceBuilds {
		jobIDs = append(jobIDs, build.JobID)
	}
	jobs, err := api.Database.GetProvisionerJobsByIDsWithQueuePosition(ctx, database.GetProvisionerJobsByIDsWithQueuePositionParams{
		IDs:             jobIDs,
		StaleIntervalMS: provisionerdserver.StaleInterval.Milliseconds(),
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return workspaceBuildsData{}, xerrors.Errorf("get provisioner jobs: %w", err)
	}
	pendingJobIDs := []uuid.UUID{}
	for _, job := range jobs {
		if job.ProvisionerJob.JobStatus == database.ProvisionerJobStatusPending {
			pendingJobIDs = append(pendingJobIDs, job.ProvisionerJob.ID)
		}
	}

	pendingJobProvisioners, err := api.Database.GetEligibleProvisionerDaemonsByProvisionerJobIDs(ctx, pendingJobIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return workspaceBuildsData{}, xerrors.Errorf("get provisioner daemons: %w", err)
	}

	templateVersionIDs := make([]uuid.UUID, 0, len(workspaceBuilds))
	for _, build := range workspaceBuilds {
		templateVersionIDs = append(templateVersionIDs, build.TemplateVersionID)
	}

	// nolint:gocritic // Getting template versions by ID is a system function.
	templateVersions, err := api.Database.GetTemplateVersionsByIDs(dbauthz.AsSystemRestricted(ctx), templateVersionIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return workspaceBuildsData{}, xerrors.Errorf("get template versions: %w", err)
	}

	// nolint:gocritic // Getting workspace resources by job ID is a system function.
	resources, err := api.Database.GetWorkspaceResourcesByJobIDs(dbauthz.AsSystemRestricted(ctx), jobIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return workspaceBuildsData{}, xerrors.Errorf("get workspace resources by job: %w", err)
	}

	if len(resources) == 0 {
		return workspaceBuildsData{
			jobs:               jobs,
			templateVersions:   templateVersions,
			provisionerDaemons: pendingJobProvisioners,
		}, nil
	}

	resourceIDs := make([]uuid.UUID, 0)
	for _, resource := range resources {
		resourceIDs = append(resourceIDs, resource.ID)
	}

	// nolint:gocritic // Getting workspace resource metadata by resource ID is a system function.
	metadata, err := api.Database.GetWorkspaceResourceMetadataByResourceIDs(dbauthz.AsSystemRestricted(ctx), resourceIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return workspaceBuildsData{}, xerrors.Errorf("fetching resource metadata: %w", err)
	}

	// nolint:gocritic // Getting workspace agents by resource IDs is a system function.
	agents, err := api.Database.GetWorkspaceAgentsByResourceIDs(dbauthz.AsSystemRestricted(ctx), resourceIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return workspaceBuildsData{}, xerrors.Errorf("get workspace agents: %w", err)
	}

	if len(resources) == 0 {
		return workspaceBuildsData{
			jobs:               jobs,
			templateVersions:   templateVersions,
			resources:          resources,
			metadata:           metadata,
			provisionerDaemons: pendingJobProvisioners,
		}, nil
	}

	agentIDs := make([]uuid.UUID, 0)
	for _, agent := range agents {
		agentIDs = append(agentIDs, agent.ID)
	}

	var (
		apps       []database.WorkspaceApp
		scripts    []database.WorkspaceAgentScript
		logSources []database.WorkspaceAgentLogSource
	)

	var eg errgroup.Group
	eg.Go(func() (err error) {
		// nolint:gocritic // Getting workspace apps by agent IDs is a system function.
		apps, err = api.Database.GetWorkspaceAppsByAgentIDs(dbauthz.AsSystemRestricted(ctx), agentIDs)
		return err
	})
	eg.Go(func() (err error) {
		// nolint:gocritic // Getting workspace scripts by agent IDs is a system function.
		scripts, err = api.Database.GetWorkspaceAgentScriptsByAgentIDs(dbauthz.AsSystemRestricted(ctx), agentIDs)
		return err
	})
	eg.Go(func() error {
		// nolint:gocritic // Getting workspace agent log sources by agent IDs is a system function.
		logSources, err = api.Database.GetWorkspaceAgentLogSourcesByAgentIDs(dbauthz.AsSystemRestricted(ctx), agentIDs)
		return err
	})
	err = eg.Wait()
	if err != nil {
		return workspaceBuildsData{}, err
	}

	appIDs := make([]uuid.UUID, 0)
	for _, app := range apps {
		appIDs = append(appIDs, app.ID)
	}

	// nolint:gocritic // Getting workspace app statuses by app IDs is a system function.
	statuses, err := api.Database.GetWorkspaceAppStatusesByAppIDs(dbauthz.AsSystemRestricted(ctx), appIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return workspaceBuildsData{}, xerrors.Errorf("get workspace app statuses: %w", err)
	}

	return workspaceBuildsData{
		jobs:               jobs,
		templateVersions:   templateVersions,
		resources:          resources,
		metadata:           metadata,
		agents:             agents,
		apps:               apps,
		appStatuses:        statuses,
		scripts:            scripts,
		logSources:         logSources,
		provisionerDaemons: pendingJobProvisioners,
	}, nil
}

func (api *API) convertWorkspaceBuilds(
	workspaceBuilds []database.WorkspaceBuild,
	workspaces []database.Workspace,
	jobs []database.GetProvisionerJobsByIDsWithQueuePositionRow,
	workspaceResources []database.WorkspaceResource,
	resourceMetadata []database.WorkspaceResourceMetadatum,
	resourceAgents []database.WorkspaceAgent,
	agentApps []database.WorkspaceApp,
	agentAppStatuses []database.WorkspaceAppStatus,
	agentScripts []database.WorkspaceAgentScript,
	agentLogSources []database.WorkspaceAgentLogSource,
	templateVersions []database.TemplateVersion,
	provisionerDaemons []database.GetEligibleProvisionerDaemonsByProvisionerJobIDsRow,
) ([]codersdk.WorkspaceBuild, error) {
	workspaceByID := map[uuid.UUID]database.Workspace{}
	for _, workspace := range workspaces {
		workspaceByID[workspace.ID] = workspace
	}
	jobByID := map[uuid.UUID]database.GetProvisionerJobsByIDsWithQueuePositionRow{}
	for _, job := range jobs {
		jobByID[job.ProvisionerJob.ID] = job
	}
	templateVersionByID := map[uuid.UUID]database.TemplateVersion{}
	for _, templateVersion := range templateVersions {
		templateVersionByID[templateVersion.ID] = templateVersion
	}

	// Should never be nil for API consistency
	apiBuilds := []codersdk.WorkspaceBuild{}
	for _, build := range workspaceBuilds {
		job, exists := jobByID[build.JobID]
		if !exists {
			return nil, xerrors.New("build job not found")
		}
		workspace, exists := workspaceByID[build.WorkspaceID]
		if !exists {
			return nil, xerrors.New("workspace not found")
		}
		templateVersion, exists := templateVersionByID[build.TemplateVersionID]
		if !exists {
			return nil, xerrors.New("template version not found")
		}

		apiBuild, err := api.convertWorkspaceBuild(
			build,
			workspace,
			job,
			workspaceResources,
			resourceMetadata,
			resourceAgents,
			agentApps,
			agentAppStatuses,
			agentScripts,
			agentLogSources,
			templateVersion,
			provisionerDaemons,
		)
		if err != nil {
			return nil, xerrors.Errorf("converting workspace build: %w", err)
		}

		apiBuilds = append(apiBuilds, apiBuild)
	}

	return apiBuilds, nil
}

func (api *API) convertWorkspaceBuild(
	build database.WorkspaceBuild,
	workspace database.Workspace,
	job database.GetProvisionerJobsByIDsWithQueuePositionRow,
	workspaceResources []database.WorkspaceResource,
	resourceMetadata []database.WorkspaceResourceMetadatum,
	resourceAgents []database.WorkspaceAgent,
	agentApps []database.WorkspaceApp,
	agentAppStatuses []database.WorkspaceAppStatus,
	agentScripts []database.WorkspaceAgentScript,
	agentLogSources []database.WorkspaceAgentLogSource,
	templateVersion database.TemplateVersion,
	provisionerDaemons []database.GetEligibleProvisionerDaemonsByProvisionerJobIDsRow,
) (codersdk.WorkspaceBuild, error) {
	resourcesByJobID := map[uuid.UUID][]database.WorkspaceResource{}
	for _, resource := range workspaceResources {
		resourcesByJobID[resource.JobID] = append(resourcesByJobID[resource.JobID], resource)
	}
	metadataByResourceID := map[uuid.UUID][]database.WorkspaceResourceMetadatum{}
	for _, metadata := range resourceMetadata {
		metadataByResourceID[metadata.WorkspaceResourceID] = append(metadataByResourceID[metadata.WorkspaceResourceID], metadata)
	}
	agentsByResourceID := map[uuid.UUID][]database.WorkspaceAgent{}
	for _, agent := range resourceAgents {
		agentsByResourceID[agent.ResourceID] = append(agentsByResourceID[agent.ResourceID], agent)
	}
	appsByAgentID := map[uuid.UUID][]database.WorkspaceApp{}
	for _, app := range agentApps {
		appsByAgentID[app.AgentID] = append(appsByAgentID[app.AgentID], app)
	}
	scriptsByAgentID := map[uuid.UUID][]database.WorkspaceAgentScript{}
	for _, script := range agentScripts {
		scriptsByAgentID[script.WorkspaceAgentID] = append(scriptsByAgentID[script.WorkspaceAgentID], script)
	}
	logSourcesByAgentID := map[uuid.UUID][]database.WorkspaceAgentLogSource{}
	for _, logSource := range agentLogSources {
		logSourcesByAgentID[logSource.WorkspaceAgentID] = append(logSourcesByAgentID[logSource.WorkspaceAgentID], logSource)
	}
	provisionerDaemonsForThisWorkspaceBuild := []database.ProvisionerDaemon{}
	for _, provisionerDaemon := range provisionerDaemons {
		if provisionerDaemon.JobID != job.ProvisionerJob.ID {
			continue
		}
		provisionerDaemonsForThisWorkspaceBuild = append(provisionerDaemonsForThisWorkspaceBuild, provisionerDaemon.ProvisionerDaemon)
	}
	matchedProvisioners := db2sdk.MatchedProvisioners(provisionerDaemonsForThisWorkspaceBuild, job.ProvisionerJob.CreatedAt, provisionerdserver.StaleInterval)
	statusesByAgentID := map[uuid.UUID][]database.WorkspaceAppStatus{}
	for _, status := range agentAppStatuses {
		statusesByAgentID[status.AgentID] = append(statusesByAgentID[status.AgentID], status)
	}

	resources := resourcesByJobID[job.ProvisionerJob.ID]
	apiResources := make([]codersdk.WorkspaceResource, 0)
	resourceAgentsMinOrder := map[uuid.UUID]int32{} // map[resource.ID]minOrder
	for _, resource := range resources {
		agents := agentsByResourceID[resource.ID]
		sort.Slice(agents, func(i, j int) bool {
			if agents[i].DisplayOrder != agents[j].DisplayOrder {
				return agents[i].DisplayOrder < agents[j].DisplayOrder
			}
			return agents[i].Name < agents[j].Name
		})

		apiAgents := make([]codersdk.WorkspaceAgent, 0)
		resourceAgentsMinOrder[resource.ID] = math.MaxInt32

		for _, agent := range agents {
			resourceAgentsMinOrder[resource.ID] = min(resourceAgentsMinOrder[resource.ID], agent.DisplayOrder)

			apps := appsByAgentID[agent.ID]
			scripts := scriptsByAgentID[agent.ID]
			statuses := statusesByAgentID[agent.ID]
			logSources := logSourcesByAgentID[agent.ID]
			apiAgent, err := db2sdk.WorkspaceAgent(
				api.DERPMap(), *api.TailnetCoordinator.Load(), agent, db2sdk.Apps(apps, statuses, agent, workspace.OwnerUsername, workspace), convertScripts(scripts), convertLogSources(logSources), api.AgentInactiveDisconnectTimeout,
				api.DeploymentValues.AgentFallbackTroubleshootingURL.String(),
			)
			if err != nil {
				return codersdk.WorkspaceBuild{}, xerrors.Errorf("converting workspace agent: %w", err)
			}
			apiAgents = append(apiAgents, apiAgent)
		}
		metadata := append(make([]database.WorkspaceResourceMetadatum, 0), metadataByResourceID[resource.ID]...)
		apiResources = append(apiResources, convertWorkspaceResource(resource, apiAgents, metadata))
	}
	sort.Slice(apiResources, func(i, j int) bool {
		orderI := resourceAgentsMinOrder[apiResources[i].ID]
		orderJ := resourceAgentsMinOrder[apiResources[j].ID]
		if orderI != orderJ {
			return orderI < orderJ
		}
		return apiResources[i].Name < apiResources[j].Name
	})

	var presetID *uuid.UUID
	if build.TemplateVersionPresetID.Valid {
		presetID = &build.TemplateVersionPresetID.UUID
	}
	var hasAITask *bool
	if build.HasAITask.Valid {
		hasAITask = &build.HasAITask.Bool
	}
	var aiTasksSidebarAppID *uuid.UUID
	if build.AITaskSidebarAppID.Valid {
		aiTasksSidebarAppID = &build.AITaskSidebarAppID.UUID
	}

	apiJob := convertProvisionerJob(job)
	transition := codersdk.WorkspaceTransition(build.Transition)
	return codersdk.WorkspaceBuild{
		ID:                      build.ID,
		CreatedAt:               build.CreatedAt,
		UpdatedAt:               build.UpdatedAt,
		WorkspaceOwnerID:        workspace.OwnerID,
		WorkspaceOwnerName:      workspace.OwnerUsername,
		WorkspaceOwnerAvatarURL: workspace.OwnerAvatarUrl,
		WorkspaceID:             build.WorkspaceID,
		WorkspaceName:           workspace.Name,
		TemplateVersionID:       build.TemplateVersionID,
		TemplateVersionName:     templateVersion.Name,
		BuildNumber:             build.BuildNumber,
		Transition:              transition,
		InitiatorID:             build.InitiatorID,
		InitiatorUsername:       build.InitiatorByUsername,
		Job:                     apiJob,
		Deadline:                codersdk.NewNullTime(build.Deadline, !build.Deadline.IsZero()),
		MaxDeadline:             codersdk.NewNullTime(build.MaxDeadline, !build.MaxDeadline.IsZero()),
		Reason:                  codersdk.BuildReason(build.Reason),
		Resources:               apiResources,
		Status:                  codersdk.ConvertWorkspaceStatus(apiJob.Status, transition),
		DailyCost:               build.DailyCost,
		MatchedProvisioners:     &matchedProvisioners,
		TemplateVersionPresetID: presetID,
		HasAITask:               hasAITask,
		AITaskSidebarAppID:      aiTasksSidebarAppID,
	}, nil
}

func convertWorkspaceResource(resource database.WorkspaceResource, agents []codersdk.WorkspaceAgent, metadata []database.WorkspaceResourceMetadatum) codersdk.WorkspaceResource {
	var convertedMetadata []codersdk.WorkspaceResourceMetadata
	for _, field := range metadata {
		convertedMetadata = append(convertedMetadata, codersdk.WorkspaceResourceMetadata{
			Key:       field.Key,
			Value:     field.Value.String,
			Sensitive: field.Sensitive,
		})
	}

	return codersdk.WorkspaceResource{
		ID:         resource.ID,
		CreatedAt:  resource.CreatedAt,
		JobID:      resource.JobID,
		Transition: codersdk.WorkspaceTransition(resource.Transition),
		Type:       resource.Type,
		Name:       resource.Name,
		Hide:       resource.Hide,
		Icon:       resource.Icon,
		Agents:     agents,
		Metadata:   convertedMetadata,
		DailyCost:  resource.DailyCost,
	}
}

func (api *API) buildTimings(ctx context.Context, build database.WorkspaceBuild) (codersdk.WorkspaceBuildTimings, error) {
	provisionerTimings, err := api.Database.GetProvisionerJobTimingsByJobID(ctx, build.JobID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return codersdk.WorkspaceBuildTimings{}, xerrors.Errorf("fetching provisioner job timings: %w", err)
	}

	//nolint:gocritic // Already checked if the build can be fetched.
	agentScriptTimings, err := api.Database.GetWorkspaceAgentScriptTimingsByBuildID(dbauthz.AsSystemRestricted(ctx), build.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return codersdk.WorkspaceBuildTimings{}, xerrors.Errorf("fetching workspace agent script timings: %w", err)
	}

	resources, err := api.Database.GetWorkspaceResourcesByJobID(ctx, build.JobID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return codersdk.WorkspaceBuildTimings{}, xerrors.Errorf("fetching workspace resources: %w", err)
	}
	resourceIDs := make([]uuid.UUID, 0, len(resources))
	for _, resource := range resources {
		resourceIDs = append(resourceIDs, resource.ID)
	}
	//nolint:gocritic // Already checked if the build can be fetched.
	agents, err := api.Database.GetWorkspaceAgentsByResourceIDs(dbauthz.AsSystemRestricted(ctx), resourceIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return codersdk.WorkspaceBuildTimings{}, xerrors.Errorf("fetching workspace agents: %w", err)
	}

	res := codersdk.WorkspaceBuildTimings{
		ProvisionerTimings:     make([]codersdk.ProvisionerTiming, 0, len(provisionerTimings)),
		AgentScriptTimings:     make([]codersdk.AgentScriptTiming, 0, len(agentScriptTimings)),
		AgentConnectionTimings: make([]codersdk.AgentConnectionTiming, 0, len(agents)),
	}

	for _, t := range provisionerTimings {
		// Ref: #15432: agent script timings must not have a zero start or end time.
		if t.StartedAt.IsZero() || t.EndedAt.IsZero() {
			api.Logger.Debug(ctx, "ignoring provisioner timing with zero start or end time",
				slog.F("workspace_id", build.WorkspaceID),
				slog.F("workspace_build_id", build.ID),
				slog.F("provisioner_job_id", t.JobID),
			)
			continue
		}

		res.ProvisionerTimings = append(res.ProvisionerTimings, codersdk.ProvisionerTiming{
			JobID:     t.JobID,
			Stage:     codersdk.TimingStage(t.Stage),
			Source:    t.Source,
			Action:    t.Action,
			Resource:  t.Resource,
			StartedAt: t.StartedAt,
			EndedAt:   t.EndedAt,
		})
	}
	for _, t := range agentScriptTimings {
		// Ref: #15432: agent script timings must not have a zero start or end time.
		if t.StartedAt.IsZero() || t.EndedAt.IsZero() {
			api.Logger.Debug(ctx, "ignoring agent script timing with zero start or end time",
				slog.F("workspace_id", build.WorkspaceID),
				slog.F("workspace_agent_id", t.WorkspaceAgentID),
				slog.F("workspace_build_id", build.ID),
				slog.F("workspace_agent_script_id", t.ScriptID),
			)
			continue
		}

		res.AgentScriptTimings = append(res.AgentScriptTimings, codersdk.AgentScriptTiming{
			StartedAt:          t.StartedAt,
			EndedAt:            t.EndedAt,
			ExitCode:           t.ExitCode,
			Stage:              codersdk.TimingStage(t.Stage),
			Status:             string(t.Status),
			DisplayName:        t.DisplayName,
			WorkspaceAgentID:   t.WorkspaceAgentID.String(),
			WorkspaceAgentName: t.WorkspaceAgentName,
		})
	}
	for _, agent := range agents {
		if agent.FirstConnectedAt.Time.IsZero() {
			api.Logger.Debug(ctx, "ignoring agent connection timing with zero first connected time",
				slog.F("workspace_id", build.WorkspaceID),
				slog.F("workspace_agent_id", agent.ID),
				slog.F("workspace_build_id", build.ID),
			)
			continue
		}
		res.AgentConnectionTimings = append(res.AgentConnectionTimings, codersdk.AgentConnectionTiming{
			WorkspaceAgentID:   agent.ID.String(),
			WorkspaceAgentName: agent.Name,
			StartedAt:          agent.CreatedAt,
			Stage:              codersdk.TimingStageConnect,
			EndedAt:            agent.FirstConnectedAt.Time,
		})
	}

	return res, nil
}
