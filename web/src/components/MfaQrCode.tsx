import { useEffect, useState } from 'react'
import QRCode from 'qrcode'

type MfaQrCodeProps = {
  otpauthUrl: string
  size?: number
}

/** Renders a QR code locally — the secret never leaves the browser. */
export function MfaQrCode({ otpauthUrl, size = 180 }: MfaQrCodeProps) {
  const [dataUrl, setDataUrl] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setDataUrl(null)
    setError(null)

    void QRCode.toDataURL(otpauthUrl, {
      width: size,
      margin: 1,
      errorCorrectionLevel: 'M',
    })
      .then((url) => {
        if (!cancelled) setDataUrl(url)
      })
      .catch(() => {
        if (!cancelled) setError('Could not render QR code')
      })

    return () => {
      cancelled = true
    }
  }, [otpauthUrl, size])

  if (error) {
    return <p className="text-sm text-amber-400">{error}</p>
  }

  if (!dataUrl) {
    return <p className="text-sm text-slate-400">Generating QR code…</p>
  }

  return (
    <img
      src={dataUrl}
      alt="MFA QR code"
      className="rounded-md border border-slate-700 bg-white p-2"
      width={size}
      height={size}
    />
  )
}
