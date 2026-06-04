// @ts-nocheck
import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute, redirect } from '@tanstack/react-router'
import { toast } from 'sonner'
import { Megaphone, Plus, Pencil, Trash2 } from 'lucide-react'
import { Topbar } from '@/components/layout/topbar'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { api } from '@/lib/api'
import { handleServerError } from '@/lib/handle-server-error'
import { profileQueryFn } from '@/lib/profile'
import { useAuthStore } from '@/stores/auth-store'

export const Route = createFileRoute('/announcements')({
  beforeLoad: () => {
    const token = useAuthStore.getState().auth.accessToken
    if (!token) throw redirect({ to: '/' })
  },
  component: AnnouncementsPage,
})

const emptyForm = () => ({
  title: '',
  content: '',
  type: 'info',
  is_active: true,
  starts_at: '',
  expires_at: '',
})

function AnnouncementsPage() {
  const queryClient = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState(null)
  const [form, setForm] = useState(emptyForm())
  const [deleteId, setDeleteId] = useState(null)

  const { data: profile } = useQuery({
    queryKey: ['profile'],
    queryFn: profileQueryFn,
  })
  const isAdmin = Boolean(profile?.is_admin)

  const { data: announcements = [], isLoading } = useQuery({
    queryKey: ['announcements'],
    queryFn: async () => {
      const res = await api.get('/api/admin/announcements')
      return res.data?.announcements || []
    },
    enabled: isAdmin,
  })

  const saveMutation = useMutation({
    mutationFn: async () => {
      const payload = {
        title: form.title.trim(),
        content: form.content.trim(),
        type: form.type,
        is_active: form.is_active,
        starts_at: form.starts_at.trim()
          ? new Date(form.starts_at).toISOString()
          : null,
        expires_at: form.expires_at.trim()
          ? new Date(form.expires_at).toISOString()
          : null,
      }
      if (editing?.id) {
        return api.put(`/api/admin/announcements/${editing.id}`, payload)
      }
      return api.post('/api/admin/announcements', payload)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['announcements'] })
      queryClient.invalidateQueries({ queryKey: ['user-announcements'] })
      toast.success(editing ? '公告已更新' : '公告已发布')
      setDialogOpen(false)
      setEditing(null)
      setForm(emptyForm())
    },
    onError: handleServerError,
  })

  const deleteMutation = useMutation({
    mutationFn: async (id) => api.delete(`/api/admin/announcements/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['announcements'] })
      queryClient.invalidateQueries({ queryKey: ['user-announcements'] })
      toast.success('公告已删除')
      setDeleteId(null)
    },
    onError: handleServerError,
  })

  const typeBadge = useMemo(
    () => ({
      info: 'bg-blue-500/10 text-blue-700 dark:text-blue-400',
      warning: 'bg-amber-500/10 text-amber-700 dark:text-amber-400',
      critical: 'bg-red-500/10 text-red-700 dark:text-red-400',
    }),
    []
  )

  const openCreate = () => {
    setEditing(null)
    setForm(emptyForm())
    setDialogOpen(true)
  }

  const openEdit = (item) => {
    setEditing(item)
    setForm({
      title: item.title,
      content: item.content,
      type: item.type || 'info',
      is_active: item.is_active,
      starts_at: item.starts_at ? item.starts_at.slice(0, 16) : '',
      expires_at: item.expires_at ? item.expires_at.slice(0, 16) : '',
    })
    setDialogOpen(true)
  }

  if (!isAdmin) {
    return (
      <>
        <Topbar />
        <main className='mx-auto flex w-full max-w-3xl flex-col items-center justify-center gap-4 px-4 py-20 text-center sm:px-6 pt-24'>
          <p className='text-muted-foreground'>仅管理员可管理公告</p>
        </main>
      </>
    )
  }

  return (
    <>
      <Topbar />
      <main className='mx-auto w-full max-w-4xl px-4 py-8 sm:px-6 pt-24'>
        <Card>
          <CardHeader className='flex flex-row items-start justify-between gap-4'>
            <div>
              <CardTitle className='flex items-center gap-2'>
                <Megaphone className='size-5' />
                公告管理
              </CardTitle>
              <CardDescription className='mt-1'>
                向所有登录用户展示系统公告，用户可在当前会话中关闭单条公告。
              </CardDescription>
            </div>
            <Button onClick={openCreate}>
              <Plus className='size-4 mr-1' />
              新建公告
            </Button>
          </CardHeader>
          <CardContent className='space-y-3'>
            {isLoading && <p className='text-sm text-muted-foreground'>加载中…</p>}
            {!isLoading && announcements.length === 0 && (
              <p className='text-sm text-muted-foreground'>暂无公告，点击「新建公告」发布第一条。</p>
            )}
            {announcements.map((item) => (
              <div
                key={item.id}
                className='flex flex-col gap-2 rounded-lg border p-4 sm:flex-row sm:items-start sm:justify-between'
              >
                <div className='min-w-0 flex-1 space-y-1'>
                  <div className='flex flex-wrap items-center gap-2'>
                    <span className='font-medium'>{item.title}</span>
                    <Badge variant='outline' className={typeBadge[item.type] || typeBadge.info}>
                      {item.type}
                    </Badge>
                    <Badge variant={item.is_active ? 'default' : 'secondary'}>
                      {item.is_active ? '已启用' : '已停用'}
                    </Badge>
                  </div>
                  <p className='text-sm text-muted-foreground whitespace-pre-wrap'>{item.content}</p>
                  <p className='text-xs text-muted-foreground'>
                    更新于 {new Date(item.updated_at).toLocaleString('zh-CN')}
                    {item.created_by ? ` · ${item.created_by}` : ''}
                  </p>
                </div>
                <div className='flex shrink-0 gap-2'>
                  <Button variant='outline' size='sm' onClick={() => openEdit(item)}>
                    <Pencil className='size-3.5 mr-1' />
                    编辑
                  </Button>
                  <Button variant='outline' size='sm' onClick={() => setDeleteId(item.id)}>
                    <Trash2 className='size-3.5 mr-1' />
                    删除
                  </Button>
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
      </main>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className='max-w-lg'>
          <DialogHeader>
            <DialogTitle>{editing ? '编辑公告' : '新建公告'}</DialogTitle>
          </DialogHeader>
          <div className='space-y-4 py-2'>
            <div className='space-y-2'>
              <Label htmlFor='ann-title'>标题</Label>
              <Input
                id='ann-title'
                value={form.title}
                onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
                placeholder='公告标题'
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='ann-content'>内容</Label>
              <Textarea
                id='ann-content'
                rows={5}
                value={form.content}
                onChange={(e) => setForm((f) => ({ ...f, content: e.target.value }))}
                placeholder='公告正文，支持多行'
              />
            </div>
            <div className='space-y-2'>
              <Label>类型</Label>
              <Select value={form.type} onValueChange={(v) => setForm((f) => ({ ...f, type: v }))}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='info'>信息</SelectItem>
                  <SelectItem value='warning'>警告</SelectItem>
                  <SelectItem value='critical'>重要</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className='flex items-center gap-2'>
              <Switch
                id='ann-active'
                checked={form.is_active}
                onCheckedChange={(v) => setForm((f) => ({ ...f, is_active: v }))}
              />
              <Label htmlFor='ann-active'>立即启用</Label>
            </div>
            <div className='grid gap-3 sm:grid-cols-2'>
              <div className='space-y-2'>
                <Label htmlFor='ann-start'>生效时间（可选）</Label>
                <Input
                  id='ann-start'
                  type='datetime-local'
                  value={form.starts_at}
                  onChange={(e) => setForm((f) => ({ ...f, starts_at: e.target.value }))}
                />
              </div>
              <div className='space-y-2'>
                <Label htmlFor='ann-end'>过期时间（可选）</Label>
                <Input
                  id='ann-end'
                  type='datetime-local'
                  value={form.expires_at}
                  onChange={(e) => setForm((f) => ({ ...f, expires_at: e.target.value }))}
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setDialogOpen(false)}>
              取消
            </Button>
            <Button
              onClick={() => saveMutation.mutate()}
              disabled={!form.title.trim() || !form.content.trim() || saveMutation.isPending}
            >
              {saveMutation.isPending ? '保存中…' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteId != null} onOpenChange={(o) => !o && setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>删除公告</DialogTitle>
          </DialogHeader>
          <p className='text-sm text-muted-foreground'>删除后所有用户将不再看到该公告，确定继续？</p>
          <DialogFooter>
            <Button variant='outline' onClick={() => setDeleteId(null)}>
              取消
            </Button>
            <Button
              variant='destructive'
              onClick={() => deleteMutation.mutate(deleteId)}
              disabled={deleteMutation.isPending}
            >
              删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
