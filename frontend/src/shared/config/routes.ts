export const ROUTES = {
  onboarding: '/onboarding',
  feed: '/',
  incident: (id = ':id') => `/incidents/${id}`,
  incidentCreate: '/incidents/new',
  houseIncidents: '/house/incidents',
  event: (id = ':id') => `/events/${id}`,
  dispatcher: '/uk-admin',
  dispatcherEventCreate: '/uk-admin/events/new',
  stats: '/house/stats',
} as const
