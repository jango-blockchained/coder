import { type CSSObject, useTheme } from "@emotion/react";
import { Avatar } from "components/Avatar/Avatar";
import type { FC, ReactNode } from "react";

type AvatarCardProps = {
	header: string;
	imgUrl: string;
	subtitle?: ReactNode;
	maxWidth?: number | "none";
};

export const AvatarCard: FC<AvatarCardProps> = ({
	header,
	imgUrl,
	subtitle,
	maxWidth = "none",
}) => {
	const theme = useTheme();

	return (
		<div
			css={{
				maxWidth: maxWidth === "none" ? undefined : `${maxWidth}px`,
				display: "flex",
				flexFlow: "row nowrap",
				alignItems: "center",
				border: `1px solid ${theme.palette.divider}`,
				gap: "16px",
				padding: "16px",
				borderRadius: "8px",
				cursor: "default",
			}}
		>
			{/**
			 * minWidth is necessary to ensure that the text truncation works properly
			 * with flex containers that don't have fixed width
			 *
			 * @see {@link https://css-tricks.com/flexbox-truncated-text/}
			 */}
			<div css={{ marginRight: "auto", minWidth: 0 }}>
				<h3
					// Lets users hover over truncated text to see whole thing
					title={header}
					css={[
						theme.typography.body1 as CSSObject,
						{
							lineHeight: 1.4,
							margin: 0,
							overflow: "hidden",
							whiteSpace: "nowrap",
							textOverflow: "ellipsis",
						},
					]}
				>
					{header}
				</h3>

				{subtitle && (
					<div
						css={[
							theme.typography.body2 as CSSObject,
							{ color: theme.palette.text.secondary },
						]}
					>
						{subtitle}
					</div>
				)}
			</div>

			<Avatar size="lg" src={imgUrl} fallback={header} />
		</div>
	);
};
