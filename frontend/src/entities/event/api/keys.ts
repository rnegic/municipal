export const eventKeys = {
  all: ['events'] as const,
  lists: () => [...eventKeys.all, 'list'] as const,
  list: (houseId: string) => [...eventKeys.all, 'list', houseId] as const,
}
