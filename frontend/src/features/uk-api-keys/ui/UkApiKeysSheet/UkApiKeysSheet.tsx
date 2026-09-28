import { useState, type FormEvent } from 'react'

import { Typography } from '@maxhub/max-ui'

import { describeApiError } from '@/shared/lib/api-error'
import { formatDateTime } from '@/shared/lib/date'
import { BottomSheet } from '@/shared/ui/bottom-sheet'
import { Button } from '@/shared/ui/button'
import { LoadingState } from '@/shared/ui/loading-state'
import { useApiKeysQuery, useCreateApiKeyMutation, useRevokeApiKeyMutation } from '../../api'
import { ukApiKeysTexts as texts } from '../../config/texts'
import type { ApiKey, CreatedApiKey } from '../../model/schema'
import s from './UkApiKeysSheet.module.scss'

export interface UkApiKeysSheetProps {
  open: boolean
  onClose: () => void
}

const CreatedKey = ({ created, onDone }: { created: CreatedApiKey; onDone: () => void }) => {
  const [copied, setCopied] = useState(false)

  return (
    <div className={s.created}>
      <Typography.Text variant="body-strong">{texts.createdTitle}</Typography.Text>
      <Typography.Text variant="note" color="secondary">
        {texts.createdWarning}
      </Typography.Text>
      <input
        className={s.keyValue}
        readOnly
        value={created.key}
        aria-label={created.name}
        onFocus={(event) => event.currentTarget.select()}
      />
      <div className={s.row}>
        <Button
          type="button"
          stretched
          onClick={() => {
            void navigator.clipboard?.writeText(created.key).then(() => setCopied(true))
          }}
        >
          {copied ? texts.copied : texts.copy}
        </Button>
        <Button type="button" tone="secondary" stretched onClick={onDone}>
          {texts.done}
        </Button>
      </div>
    </div>
  )
}

const KeyRow = ({ apiKey }: { apiKey: ApiKey }) => {
  const revokeMutation = useRevokeApiKeyMutation()
  const [confirming, setConfirming] = useState(false)
  const isRevoked = apiKey.revokedAt !== null

  return (
    <li className={isRevoked ? `${s.key} ${s.keyRevoked}` : s.key}>
      <div className={s.keyInfo}>
        <Typography.Text variant="body-strong">{apiKey.name}</Typography.Text>
        <Typography.Text variant="note" color="tertiary">
          uk_live_{apiKey.prefix}…
        </Typography.Text>
        <Typography.Text variant="note" color="secondary">
          {texts.created}: {formatDateTime(apiKey.createdAt)}
        </Typography.Text>
        <Typography.Text variant="note" color="secondary">
          {isRevoked && apiKey.revokedAt
            ? `${texts.revoked}: ${formatDateTime(apiKey.revokedAt)}`
            : `${texts.lastUsed}: ${apiKey.lastUsedAt ? formatDateTime(apiKey.lastUsedAt) : texts.neverUsed}`}
        </Typography.Text>
        {revokeMutation.error ? (
          <Typography.Text className={s.error} variant="note">
            {describeApiError(revokeMutation.error).description}
          </Typography.Text>
        ) : null}
      </div>
      {isRevoked ? null : confirming ? (
        <div className={s.keyActions}>
          <Button
            type="button"
            size="small"
            disabled={revokeMutation.isPending}
            onClick={() => revokeMutation.mutate(apiKey.id)}
          >
            {texts.confirmRevoke}
          </Button>
          <Button
            type="button"
            size="small"
            tone="secondary"
            disabled={revokeMutation.isPending}
            onClick={() => setConfirming(false)}
          >
            {texts.cancel}
          </Button>
        </div>
      ) : (
        <Button type="button" size="small" tone="secondary" onClick={() => setConfirming(true)}>
          {texts.revoke}
        </Button>
      )}
    </li>
  )
}

export const UkApiKeysSheet = ({ open, onClose }: UkApiKeysSheetProps) => {
  const keysQuery = useApiKeysQuery(open)
  const createMutation = useCreateApiKeyMutation()
  const [name, setName] = useState('')
  const [validationError, setValidationError] = useState<string | null>(null)
  const [created, setCreated] = useState<CreatedApiKey | null>(null)

  const close = () => {
    setCreated(null)
    setName('')
    setValidationError(null)
    createMutation.reset()
    onClose()
  }

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    if (!name.trim()) {
      setValidationError(texts.nameRequired)
      return
    }

    setValidationError(null)

    try {
      setCreated(await createMutation.mutateAsync(name.trim()))
      setName('')
    } catch {
      setValidationError(null)
    }
  }

  return (
    <BottomSheet open={open} title={texts.title} description={texts.description} onClose={close}>
      <div className={s.root}>
        {created ? (
          <CreatedKey created={created} onDone={() => setCreated(null)} />
        ) : (
          <form className={s.form} onSubmit={handleSubmit}>
            <label className={s.field}>
              <Typography.Text variant="description" color="secondary">
                {texts.nameLabel}
              </Typography.Text>
              <input
                className={s.input}
                maxLength={100}
                placeholder={texts.namePlaceholder}
                value={name}
                onChange={(event) => setName(event.target.value)}
              />
            </label>
            {validationError ? (
              <Typography.Text className={s.error} variant="note">
                {validationError}
              </Typography.Text>
            ) : null}
            {createMutation.error ? (
              <Typography.Text className={s.error} variant="note">
                {describeApiError(createMutation.error).description}
              </Typography.Text>
            ) : null}
            <Button type="submit" stretched disabled={createMutation.isPending}>
              {texts.create}
            </Button>
            <Typography.Text variant="note" color="tertiary">
              {texts.limitHint}
            </Typography.Text>
          </form>
        )}
        <section className={s.list} aria-label={texts.listTitle}>
          <Typography.Text variant="body-strong">{texts.listTitle}</Typography.Text>
          {keysQuery.isPending ? (
            <LoadingState />
          ) : keysQuery.isError ? (
            <Typography.Text className={s.error} variant="note">
              {describeApiError(keysQuery.error).description}
            </Typography.Text>
          ) : keysQuery.data.items.length === 0 ? (
            <Typography.Text variant="note" color="secondary">
              {texts.empty}
            </Typography.Text>
          ) : (
            <ul className={s.keys}>
              {[...keysQuery.data.items].reverse().map((apiKey) => (
                <KeyRow key={apiKey.id} apiKey={apiKey} />
              ))}
            </ul>
          )}
        </section>
      </div>
    </BottomSheet>
  )
}
