import * as vaultApi from '../api/vault'
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

  const { items } = await vaultApi.listVaultItems({ limit: 100 })
  for (const item of items) {
    if (item.DeletedAt) continue

    let payload: VaultItemPayload
    try {
      payload = await decryptPayload(legacyKey, base64ToBytes(item.EncryptedData))
    } catch {
      continue
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
