import { getErrorMessage } from "api/errors";
import { deleteOrganizationRole, organizationRoles } from "api/queries/roles";
import type { Role } from "api/typesGenerated";
import { DeleteDialog } from "components/Dialogs/DeleteDialog/DeleteDialog";
import { displayError, displaySuccess } from "components/GlobalSnackbar/utils";
import { Loader } from "components/Loader/Loader";
import { SettingsHeader } from "components/SettingsHeader/SettingsHeader";
import { Stack } from "components/Stack/Stack";
import { RequirePermission } from "contexts/auth/RequirePermission";
import { useFeatureVisibility } from "modules/dashboard/useFeatureVisibility";
import { useOrganizationSettings } from "modules/management/OrganizationSettingsLayout";
import { type FC, useEffect, useState } from "react";
import { Helmet } from "react-helmet-async";
import { useMutation, useQuery, useQueryClient } from "react-query";
import { useParams } from "react-router-dom";
import { pageTitle } from "utils/page";
import CustomRolesPageView from "./CustomRolesPageView";

export const CustomRolesPage: FC = () => {
	const queryClient = useQueryClient();
	const { custom_roles: isCustomRolesEnabled } = useFeatureVisibility();
	const { organization: organizationName } = useParams() as {
		organization: string;
	};
	const { organizationPermissions } = useOrganizationSettings();

	const [roleToDelete, setRoleToDelete] = useState<Role>();

	const organizationRolesQuery = useQuery(organizationRoles(organizationName));
	const builtInRoles = organizationRolesQuery.data?.filter(
		(role) => role.built_in,
	);
	const customRoles = organizationRolesQuery.data?.filter(
		(role) => !role.built_in,
	);

	const deleteRoleMutation = useMutation(
		deleteOrganizationRole(queryClient, organizationName),
	);

	useEffect(() => {
		if (organizationRolesQuery.error) {
			displayError(
				getErrorMessage(
					organizationRolesQuery.error,
					"Error loading custom roles.",
				),
			);
		}
	}, [organizationRolesQuery.error]);

	if (!organizationPermissions) {
		return <Loader />;
	}

	return (
		<RequirePermission
			isFeatureVisible={
				organizationPermissions.assignOrgRoles ||
				organizationPermissions.createOrgRoles ||
				organizationPermissions.viewOrgRoles
			}
		>
			<Helmet>
				<title>{pageTitle("Custom Roles")}</title>
			</Helmet>

			<Stack
				alignItems="baseline"
				direction="row"
				justifyContent="space-between"
			>
				<SettingsHeader
					title="Roles"
					description="Manage roles for this organization."
				/>
			</Stack>

			<CustomRolesPageView
				builtInRoles={builtInRoles}
				customRoles={customRoles}
				onDeleteRole={setRoleToDelete}
				canAssignOrgRole={organizationPermissions.assignOrgRoles}
				canCreateOrgRole={organizationPermissions.createOrgRoles}
				isCustomRolesEnabled={isCustomRolesEnabled}
			/>

			<DeleteDialog
				key={roleToDelete?.name}
				isOpen={roleToDelete !== undefined}
				confirmLoading={deleteRoleMutation.isLoading}
				name={roleToDelete?.name ?? ""}
				entity="role"
				onCancel={() => setRoleToDelete(undefined)}
				onConfirm={async () => {
					try {
						if (roleToDelete) {
							await deleteRoleMutation.mutateAsync(roleToDelete.name);
						}
						setRoleToDelete(undefined);
						await organizationRolesQuery.refetch();
						displaySuccess("Custom role deleted successfully!");
					} catch (error) {
						displayError(
							getErrorMessage(error, "Failed to delete custom role"),
						);
					}
				}}
			/>
		</RequirePermission>
	);
};

export default CustomRolesPage;
