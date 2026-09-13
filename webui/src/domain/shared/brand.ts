declare const brand: unique symbol;

/** Nominal typing helper for identifiers that share a primitive wire shape. */
export type Branded<T, Name> = T & { readonly [brand]: Name };
