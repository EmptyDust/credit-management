export const routerFuture = {
  v7_relativeSplatPath: true,
} as const;

export const routerProviderFuture = {
  v7_startTransition: true,
} as const;

export const memoryRouterFuture = {
  ...routerFuture,
  ...routerProviderFuture,
} as const;
