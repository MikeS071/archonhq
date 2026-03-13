import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import type { Snippet } from "svelte";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

type WithChild<T = unknown> = T & {
  child?: Snippet<[Record<string, unknown>]>;
};

type WithChildren<T = unknown> = T & {
  children?: Snippet;
};

export type WithoutChild<T> = T extends WithChild ? Omit<T, "child"> : T;
export type WithoutChildren<T> = T extends WithChildren ? Omit<T, "children"> : T;
export type WithoutChildrenOrChild<T> = WithoutChild<WithoutChildren<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & {
  ref?: U | null;
};
