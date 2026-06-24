import { useState, useEffect, useCallback } from 'react'
import {
  Box, Button, Container, Snackbar, Alert, Typography,
  AppBar, Toolbar, Chip, LinearProgress, Paper,
} from '@mui/material'
import AddIcon from '@mui/icons-material/Add'
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings'
import {
  getAdminFoodEntries, createAdminFoodEntry, updateAdminFoodEntry,
  deleteAdminFoodEntry, getAdminReports,
} from '../services/api'
import { clearToken } from '../lib/auth'
import { useSnack } from '../hooks/useSnack'
import ReportCards from '../components/ReportCards'
import DateRangeFilter from '../components/DateRangeFilter'
import FoodTable from '../components/FoodTable'
import FoodFormDialog from '../components/FoodFormDialog'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog'
import type { FoodEntry, AdminReport, FoodEntryPayload } from '../types/api'

function todayStr() { return new Date().toISOString().slice(0, 10) }

export default function AdminPage() {
  const [entries, setEntries] = useState<FoodEntry[]>([])
  const [report, setReport] = useState<AdminReport | null>(null)
  const [loading, setLoading] = useState(false)
  const [dateFrom, setDateFrom] = useState(todayStr)
  const [dateTo, setDateTo] = useState(todayStr)
  const [formOpen, setFormOpen] = useState(false)
  const [editEntry, setEditEntry] = useState<FoodEntry | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<FoodEntry | null>(null)
  const { snack, show, close } = useSnack()

  const loadEntries = useCallback(async (from: string, to: string) => {
    setLoading(true)
    try {
      const data = await getAdminFoodEntries(from, to)
      setEntries(data ?? [])
    } catch (e) {
      show((e as Error).message, 'error')
    } finally {
      setLoading(false)
    }
  }, [show])

  const loadReport = useCallback(async () => {
    try {
      const data = await getAdminReports()
      setReport(data)
    } catch (e) {
      const err = e as Error & { status?: number }
      if (err.status === 403) show('ไม่มีสิทธิ์เข้าถึง report', 'error')
    }
  }, [show])

  useEffect(() => {
    loadEntries(dateFrom, dateTo)
    loadReport()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  function handleApplyFilter(from: string, to: string) {
    setDateFrom(from)
    setDateTo(to)
    loadEntries(from, to)
  }

  async function handleCreate(payload: FoodEntryPayload) {
    await createAdminFoodEntry(payload)
    show('เพิ่มรายการสำเร็จ')
    loadEntries(dateFrom, dateTo)
    loadReport()
  }

  async function handleUpdate(payload: FoodEntryPayload) {
    if (!editEntry) return
    await updateAdminFoodEntry(editEntry.id, payload)
    show('แก้ไขรายการสำเร็จ')
    loadEntries(dateFrom, dateTo)
  }

  async function handleDelete() {
    if (!deleteTarget) return
    await deleteAdminFoodEntry(deleteTarget.id)
    show('ลบรายการสำเร็จ')
    loadEntries(dateFrom, dateTo)
    loadReport()
  }

  function handleLogout() { clearToken(); window.location.reload() }

  return (
    <>
      <AppBar position="static" color="secondary">
        <Toolbar>
          <AdminPanelSettingsIcon sx={{ mr: 1 }} />
          <Typography variant="h6" sx={{ fontWeight: 'bold', flexGrow: 1 }}>
            Calorie Tracker
          </Typography>
          <Chip label="Admin" color="warning" size="small" sx={{ mr: 2 }} />
          <Button color="inherit" onClick={handleLogout}>ออกจากระบบ</Button>
        </Toolbar>
      </AppBar>
      {loading && <LinearProgress color="secondary" />}

      <Container maxWidth="lg" sx={{ py: 4 }}>
        <ReportCards report={report} />

        <Paper variant="outlined" sx={{ p: { xs: 2, sm: 3 } }}>
          <Typography
            variant="subtitle2"
            sx={{ mb: 2, fontWeight: 600, color: 'text.secondary', textTransform: 'uppercase', letterSpacing: 0.5 }}
          >
            รายการอาหาร
          </Typography>
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'flex-start',
              flexWrap: 'wrap',
              gap: 1,
              mb: 2,
            }}
          >
            <DateRangeFilter onApply={handleApplyFilter} />
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => { setEditEntry(null); setFormOpen(true) }}
            >
              เพิ่มรายการ
            </Button>
          </Box>

          <FoodTable
            entries={entries}
            isAdmin
            onEdit={(e) => { setEditEntry(e); setFormOpen(true) }}
            onDelete={(e) => setDeleteTarget(e)}
          />
        </Paper>
      </Container>

      <FoodFormDialog
        open={formOpen}
        mode={editEntry ? 'edit' : 'create'}
        entry={editEntry}
        isAdmin
        onClose={() => { setFormOpen(false); setEditEntry(null) }}
        onSubmit={editEntry ? handleUpdate : handleCreate}
      />

      <DeleteConfirmDialog
        open={!!deleteTarget}
        entry={deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
      />

      <Snackbar
        open={snack.open}
        autoHideDuration={3000}
        onClose={close}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert severity={snack.severity} onClose={close}>{snack.msg}</Alert>
      </Snackbar>
    </>
  )
}
