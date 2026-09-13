import {
	ActivityIcon,
	AirplayIcon,
	BeakerIcon,
	BookOpenIcon,
	Building2Icon,
	CalendarClockIcon,
	CircleDotIcon,
	ClipboardListIcon,
	FlaskConicalIcon,
	FolderTreeIcon,
	GitBranchIcon,
	HospitalIcon,
	type LucideIcon,
	NetworkIcon,
	NotebookTabsIcon,
	PillIcon,
	PlaneIcon,
	ScrollTextIcon,
	ShieldCheckIcon,
	StethoscopeIcon,
	UserIcon,
	UsersIcon,
	ZapIcon,
} from "lucide-react";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";

const icons = {
	activity: ActivityIcon,
	airplay: AirplayIcon,
	beaker: BeakerIcon,
	"book-open": BookOpenIcon,
	"building-2": Building2Icon,
	"calendar-clock": CalendarClockIcon,
	"circle-dot": CircleDotIcon,
	"clipboard-list": ClipboardListIcon,
	"flask-conical": FlaskConicalIcon,
	"folder-tree": FolderTreeIcon,
	"git-branch": GitBranchIcon,
	hospital: HospitalIcon,
	network: NetworkIcon,
	"notebook-tabs": NotebookTabsIcon,
	pill: PillIcon,
	plane: PlaneIcon,
	"scroll-text": ScrollTextIcon,
	"shield-check": ShieldCheckIcon,
	stethoscope: StethoscopeIcon,
	user: UserIcon,
	users: UsersIcon,
	zap: ZapIcon,
} satisfies Record<string, LucideIcon>;

export type IconName = keyof typeof icons;
export const allIconNamesSorted = Object.keys(icons).toSorted() as IconName[];

const LUCIDE_PREFIX = "lucide:";

export function isLucideIconRef(
	iconUrl: string | undefined,
): iconUrl is `lucide:${string}` {
	return !!iconUrl && iconUrl.startsWith(LUCIDE_PREFIX);
}

export function isValidLucideIconName(name: string): name is IconName {
	return Object.hasOwn(icons, name);
}

export function getLucideName(iconUrl: string | undefined): string | null {
	return isLucideIconRef(iconUrl) ? iconUrl.slice(LUCIDE_PREFIX.length) : null;
}

export function toLucideIconRef(iconName: IconName): string {
	return `${LUCIDE_PREFIX}${iconName}`;
}

export function getLucideIcon(iconName: IconName): LucideIcon {
	return icons[iconName];
}

export async function getLucideSvgString(
	iconName: string,
): Promise<string | null> {
	if (!isValidLucideIconName(iconName)) {
		return null;
	}
	return renderToStaticMarkup(createElement(icons[iconName]));
}
