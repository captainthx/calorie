import { useState } from 'react'
import {
  Box, Button, Card, CardContent, Stack, TextField, Typography, Divider,
} from '@mui/material'
import LockOpenIcon from '@mui/icons-material/LockOpen'
import { setToken, getRole } from '../lib/auth'

const QUICK_TOKENS = [
  { label: 'John (User)', token: 'user-token-123' },
  { label: 'Jane (User)', token: 'user-token-456' },
  { label: 'Admin', token: 'admin-token-789' },
] as const

export default function TokenPage() {
  const [input, setInput] = useState('')
  const [error, setError] = useState('')

  function submit(token: string) {
    const trimmed = token.trim()
    if (!getRole(trimmed)) {
      setError('Token ไม่ถูกต้อง')
      return
    }
    setToken(trimmed)
    window.location.reload()
  }

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: 'grey.100',
      }}
    >
      <Card sx={{ width: 420, p: 1 }} elevation={4}>
        <CardContent>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 1 }}>
            <LockOpenIcon color="primary" sx={{ fontSize: 32 }} />
            <Typography variant="h5" sx={{ fontWeight: 'bold' }}>Calorie Tracker</Typography>
          </Box>
          <Typography variant="body2" sx={{ color: 'text.secondary', mb: 3 }}>
            ใส่ token เพื่อเข้าใช้งาน
          </Typography>

          <Stack spacing={2}>
            <TextField
              label="API Token"
              value={input}
              onChange={(e) => { setInput(e.target.value); setError('') }}
              error={!!error}
              helperText={error}
              fullWidth
              onKeyDown={(e) => e.key === 'Enter' && submit(input)}
            />
            <Button
              variant="contained"
              size="large"
              onClick={() => submit(input)}
              disabled={!input.trim()}
            >
              เข้าใช้งาน
            </Button>
          </Stack>

          <Divider sx={{ my: 3 }}>Quick Select</Divider>

          <Stack spacing={1}>
            {QUICK_TOKENS.map(({ label, token }) => (
              <Button key={token} variant="outlined" onClick={() => submit(token)} fullWidth>
                {label}
              </Button>
            ))}
          </Stack>
        </CardContent>
      </Card>
    </Box>
  )
}
