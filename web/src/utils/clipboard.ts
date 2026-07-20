export async function copyToClipboard(text: string): Promise<void> {
  if (!text) {
    throw new Error('nothing to copy')
  }
  await navigator.clipboard.writeText(text)
}

export function normalizeUrl(url: string): string {
  const trimmed = url.trim()
  if (!trimmed) return trimmed
  if (/^https?:\/\//i.test(trimmed)) return trimmed
  return `https://${trimmed}`
}
