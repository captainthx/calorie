import { useState } from 'react'

interface SnackState {
  open: boolean
  msg: string
  severity: 'success' | 'error' | 'warning' | 'info'
}

export function useSnack() {
  const [snack, setSnack] = useState<SnackState>({ open: false, msg: '', severity: 'success' })

  const show = (msg: string, severity: SnackState['severity'] = 'success') =>
    setSnack({ open: true, msg, severity })

  const close = () => setSnack((s) => ({ ...s, open: false }))

  return { snack, show, close }
}
