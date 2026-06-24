import { useState, useEffect, useCallback } from 'react'
import {
  Box, Button, Container, Snackbar, Alert, Typography,
  AppBar, Toolbar, LinearProgress, Paper,
} from '@mui/material'
import AddIcon from '@mui/icons-material/Add'
import { getFoodEntries, createFoodEntry, updateFoodEntry, deleteFoodEntry, getDailySummary } from '../services/api'
import { clearToken } from '../lib/auth'
import { useSnack } from '../hooks/useSnack'
import SummaryCards from '../components/SummaryCards'
import DateRangeFilter from '../components/DateRangeFilter'
import FoodTable from '../components/FoodTable'
import FoodFormDialog from '../components/FoodFormDialog'
import DeleteConfirmDialog from '../components/DeleteConfirmDialog'
import type { FoodEntry, DailySummary, FoodEntryPayload } from '../types/api'
import { todayStr } from '../lib/date'

export default function UserPage() {
  const [entries, setEntries] = useState<FoodEntry[]>([])
  const [summary, setSummary] = useState<DailySummary | null>(null)
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
      const data = await getFoodEntries(from, to)
      setEntries(data ?? [])
    } catch (e) {
      show((e as Error).message, 'error')
    } finally {
      setLoading(false)
    }
  }, [show])

  const loadSummary = useCallback(async () => {
    try {
      const data = await getDailySummary(todayStr())
      setSummary(data)
    } catch { /* non-fatal */ }
  }, [])

  useEffect(() => {
    loadEntries(dateFrom, dateTo)
    loadSummary()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  function handleApplyFilter(from: string, to: string) {
    setDateFrom(from)
    setDateTo(to)
    loadEntries(from, to)
  }

  async function handleCreate(payload: FoodEntryPayload) {
    await createFoodEntry(payload)
    show('เพิ่มรายการสำเร็จ')
    loadEntries(dateFrom, dateTo)
    loadSummary()
  }

  async function handleUpdate(payload: FoodEntryPayload) {
    if (!editEntry) return
    await updateFoodEntry(editEntry.id, payload)
    show('แก้ไขรายการสำเร็จ')
    loadEntries(dateFrom, dateTo)
    loadSummary()
  }

  async function handleDelete() {
    if (!deleteTarget) return
    await deleteFoodEntry(deleteTarget.id)
    show('ลบรายการสำเร็จ')
    loadEntries(dateFrom, dateTo)
    loadSummary()
  }

  function handleLogout() { clearToken(); window.location.reload() }

  return (
    <>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" sx={{ fontWeight: 'bold', flexGrow: 1 }}>
            Calorie Tracker
          </Typography>
          <Button color="inherit" onClick={handleLogout}>ออกจากระบบ</Button>
        </Toolbar>
      </AppBar>
      {loading && <LinearProgress />}

      <Container maxWidth="lg" sx={{ py: 4 }}>
        <SummaryCards summary={summary} />

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
            onEdit={(e) => { setEditEntry(e); setFormOpen(true) }}
            onDelete={(e) => setDeleteTarget(e)}
          />
        </Paper>
      </Container>

      <FoodFormDialog
        open={formOpen}
        mode={editEntry ? 'edit' : 'create'}
        entry={editEntry}
        isAdmin={false}
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
