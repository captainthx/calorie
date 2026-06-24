import { useState, useEffect } from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
} from "@mui/material";
import WarningAmberIcon from "@mui/icons-material/WarningAmber";
import type { FoodEntry } from "../types/api";

interface Props {
  open: boolean;
  entry: FoodEntry | null;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}

export default function DeleteConfirmDialog({
  open,
  entry,
  onClose,
  onConfirm,
}: Readonly<Props>) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => { if (open) setError('') }, [open]);

  async function handleConfirm() {
    setLoading(true);
    setError('');
    try {
      await onConfirm();
      onClose();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle sx={{ display: "flex", alignItems: "center", gap: 1 }}>
        <WarningAmberIcon color="warning" />
        ยืนยันการลบ
      </DialogTitle>
      <DialogContent>
        <Typography>
          ต้องการลบ <strong>{entry?.food_name}</strong> ใช่หรือไม่?
        </Typography>
        {error && (
          <Typography variant="body2" color="error" sx={{ mt: 1 }}>{error}</Typography>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={loading}>
          ยกเลิก
        </Button>
        <Button
          variant="contained"
          color="error"
          onClick={handleConfirm}
          disabled={loading}
        >
          {loading ? "กำลังลบ..." : "ลบ"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
