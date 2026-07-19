import { apiJson } from './client'
import type {
  VaultItem,
  VaultItemCreateRequest,
  VaultItemListResponse,
  VaultItemUpdateRequest,
  VersionRequest,
} from './types'

export async function listVaultItems(params?: {
  folder?: string
  item_type?: string
  tag?: string
  title?: string
  limit?: number
  offset?: number
}): Promise<VaultItemListResponse> {
  const query = new URLSearchParams()
  if (params?.folder) query.set('folder', params.folder)
  if (params?.item_type) query.set('item_type', params.item_type)
  if (params?.tag) query.set('tag', params.tag)
  if (params?.title) query.set('title', params.title)
  if (params?.limit != null) query.set('limit', String(params.limit))
  if (params?.offset != null) query.set('offset', String(params.offset))

  const suffix = query.size > 0 ? `?${query.toString()}` : ''
  return apiJson<VaultItemListResponse>(`/api/v1/vault/items${suffix}`)
}

export async function getVaultItem(id: string): Promise<VaultItem> {
  return apiJson<VaultItem>(`/api/v1/vault/items/${id}`)
}

export async function createVaultItem(body: VaultItemCreateRequest): Promise<VaultItem> {
  return apiJson<VaultItem>('/api/v1/vault/items', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export async function updateVaultItem(
  id: string,
  body: VaultItemUpdateRequest,
): Promise<VaultItem> {
  return apiJson<VaultItem>(`/api/v1/vault/items/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  })
}

export async function deleteVaultItem(id: string, body: VersionRequest): Promise<VaultItem> {
  return apiJson<VaultItem>(`/api/v1/vault/items/${id}`, {
    method: 'DELETE',
    body: JSON.stringify(body),
  })
}

export async function restoreVaultItem(id: string, body: VersionRequest): Promise<VaultItem> {
  return apiJson<VaultItem>(`/api/v1/vault/items/${id}/restore`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}
