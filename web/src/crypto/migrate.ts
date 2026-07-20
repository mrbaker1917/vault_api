import * as vaultApi from '../api/vault'
import type { VaultItem } from '../api/types'
import { decryptPayload, encryptPayload } from './blob'
import { deriveVaultKey } from './kdf'
import {
  clearLegacySalt,
  deriveDeterministicSalt,
  getPrimarySalt,
  loadLegacySalt,
  bytesEqual,
} from './salt'
import { refreshVerifier } from './verifier'
import type { VaultItemPayload } from './types'
import { base64ToBytes, bytesToBase64 } from './encoding'

export class MigrationError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'MigrationError'
  }
}

const PAGE_SIZE = 100

async function fetchAllActiveVaultItems(): Promise<VaultItem[]> {
  const all: VaultItem[] = []
  let offset = 0

  while (true) {
    const page = await vaultApi.listVaultItems({ limit: PAGE_SIZE, offset })
    for (const item of page.items) {
      if (!item.DeletedAt) {
        all.push(item)
      }
    }
    offset += page.limit
    if (offset >= page.total) {
      break
    }
  }

  return all
}

/** Re-encrypt vault items from a legacy per-browser salt to the portable deterministic salt. */
export async function migrateLegacyEncryption(
  userId: string,
  masterPassword: string,
  legacySalt: Uint8Array,
): Promise<void> {
  const deterministicSalt = await deriveDeterministicSalt(userId)
  if (bytesEqual(legacySalt, deterministicSalt)) {
    return
  }

  const legacyKey = await deriveVaultKey(masterPassword, legacySalt)
  const newKey = await deriveVaultKey(masterPassword, deterministicSalt)
  const items = await fetchAllActiveVaultItems()

  for (const item of items) {
    let payload: VaultItemPayload
    try {
      payload = await decryptPayload(legacyKey, base64ToBytes(item.EncryptedData))
    } catch {
      throw new MigrationError(
        `Could not decrypt "${item.Title || item.ID}" during migration. Legacy salt was not removed.`,
      )
    }

    const blob = await encryptPayload(newKey, payload)
    await vaultApi.updateVaultItem(item.ID, {
      encrypted_data: bytesToBase64(blob),
      item_type: item.ItemType,
      title: item.Title,
      folder: item.Folder,
      tags: item.Tags,
      version: item.Version,
    })
  }

  clearLegacySalt(userId)
  await refreshVerifier(userId, masterPassword)
}

export async function migrateIfNeeded(
  userId: string,
  masterPassword: string,
  usedSalt: Uint8Array,
): Promise<void> {
  const legacy = loadLegacySalt(userId)
  const primary = await getPrimarySalt(userId)

  if (legacy && bytesEqual(usedSalt, legacy) && !bytesEqual(legacy, primary)) {
    await migrateLegacyEncryption(userId, masterPassword, legacy)
  }
}
