// @ts-nocheck
import { useQuery } from '@tanstack/react-query'
import { X, Info, AlertTriangle, AlertCircle } from 'lucide-react'
import { useState } from 'react'
import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'

const DISMISS_KEY = 'mmw-dismissed-announcements'

function loadDismissed(): Set<number> {
  try {
    const raw = sessionStorage.getItem(DISMISS_KEY)
    if (!raw) return new Set()
    return new Set(JSON.parse(raw))
  } catch {
    return new Set()
  }
}

function saveDismissed(ids: Set<number>) {
  sessionStorage.setItem(DISMISS_KEY, JSON.stringify([...ids]))
}

const typeIcon = {
  info: Info,
  warning: AlertTriangle,
  critical: AlertCircle,
}

const typeClass = {
  info: 'border-blue-500/30 bg-blue-500/5',
  warning: 'border-amber-500/40 bg-amber-500/10',
  critical: 'border-red-500/40 bg-red-500/10',
}

export function AnnouncementBanner() {
  const { auth } = useAuthStore()
  const [dismissed, setDismissed] = useState<Set<number>>(() => loadDismissed())

  const { data } = useQuery({
    queryKey: ['user-announcements'],
    queryFn: async () => {
      const res = await api.get('/api/user/announcements')
      return res.data?.announcements || []
    },
    enabled: Boolean(auth.accessToken),
    staleTime: 60_000,
  })

  if (!auth.accessToken || !data?.length) return null

  const visible = data.filter((a) => !dismissed.has(a.id))
  if (!visible.length) return null

  const dismiss = (id: number) => {
    const next = new Set(dismissed)
    next.add(id)
    setDismissed(next)
    saveDismissed(next)
  }

  return (
    <div className='w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80'>
      <div className='mx-auto flex w-full max-w-6xl flex-col gap-2 px-4 py-2 sm:px-6'>
        {visible.map((item) => {
          const Icon = typeIcon[item.type] || Info
          return (
            <Alert
              key={item.id}
              className={`relative pr-10 ${typeClass[item.type] || typeClass.info}`}
            >
              <Icon className='size-4' />
              <AlertTitle>{item.title}</AlertTitle>
              <AlertDescription className='whitespace-pre-wrap text-sm'>
                {item.content}
              </AlertDescription>
              <Button
                variant='ghost'
                size='icon'
                className='absolute right-1 top-1 size-7'
                onClick={() => dismiss(item.id)}
                aria-label='关闭公告'
              >
                <X className='size-4' />
              </Button>
            </Alert>
          )
        })}
      </div>
    </div>
  )
}
