export const incidentKeys = {
  all: ['incidents'] as const,
  lists: () => [...incidentKeys.all, 'list'] as const,
  list: (houseId: string) => [...incidentKeys.all, 'list', houseId] as const,
  details: () => [...incidentKeys.all, 'byId'] as const,
  byId: (incidentId: string) => [...incidentKeys.all, 'byId', incidentId] as const,
  requests: (houseId: string, offset: number, limit: number) =>
    [...incidentKeys.all, 'requests', houseId, { offset, limit }] as const,
  ukQueue: (offset: number, limit: number) => [...incidentKeys.all, 'ukQueue', { offset, limit }] as const,
}
