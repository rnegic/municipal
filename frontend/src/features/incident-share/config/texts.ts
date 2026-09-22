export interface IncidentShareMessage {
  severity: string
  title: string
  status: string
  description: string
}

export const incidentShareTexts = {
  action: 'Поделиться в домовом чате',
  message: ({ severity, title, status, description }: IncidentShareMessage) =>
    `${severity}: ${title}\nСтатус: ${status}\n${description}`,
} as const
