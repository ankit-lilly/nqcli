import { atom, useAtomValue } from "jotai";

export type TextTransformer = (text: string) => string;

/** A no-op transform: text passes through unchanged. */
export const identityTransform: TextTransformer = (text) => text;

export const textTransformSelector = atom<TextTransformer>(
	() => identityTransform,
);

function useTextTransform() {
	return useAtomValue(textTransformSelector);
}

export default useTextTransform;
