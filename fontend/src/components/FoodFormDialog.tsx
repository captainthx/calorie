import { useState, useEffect } from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Stack,
  MenuItem,
  Typography,
} from "@mui/material";
import type { FoodEntry, FoodEntryPayload } from "../types/api";
import { todayStr } from "../lib/date";

const OWNER_OPTIONS = [
  { label: "John", value: 1 },
  { label: "Jane", value: 2 },
];

interface FormState {
  food_name: string;
  calories: string;
  price: string;
  entry_date: string;
  user_id: string;
}

function emptyForm(): FormState {
  return {
    food_name: "",
    calories: "",
    price: "",
    entry_date: todayStr(),
    user_id: "",
  };
}

interface Props {
  open: boolean;
  mode: "create" | "edit";
  entry: FoodEntry | null;
  isAdmin: boolean;
  onClose: () => void;
  onSubmit: (payload: FoodEntryPayload) => Promise<void>;
}

export default function FoodFormDialog({
  open,
  mode,
  entry,
  isAdmin,
  onClose,
  onSubmit,
}: Readonly<Props>) {
  const [form, setForm] = useState<FormState>(emptyForm);
  const [errors, setErrors] = useState<Partial<FormState>>({});
  const [loading, setLoading] = useState(false);
  const [submitError, setSubmitError] = useState('');

  useEffect(() => {
    if (open) {
      setErrors({});
      setLoading(false);
      setSubmitError('');
      if (mode === "edit" && entry) {
        setForm({
          food_name: entry.food_name ?? "",
          calories: String(entry.calories ?? ""),
          price: String(entry.price ?? ""),
          entry_date: entry.entry_date
            ? entry.entry_date.slice(0, 10)
            : todayStr(),
          user_id: String(entry.user_id ?? ""),
        });
      } else {
        setForm(emptyForm());
      }
    }
  }, [open, mode, entry]);

  function set(field: keyof FormState, value: string) {
    setForm((f) => ({ ...f, [field]: value }));
    setErrors((e) => ({ ...e, [field]: "" }));
  }

  function validate(): Partial<FormState> {
    const e: Partial<FormState> = {};
    if (!form.food_name.trim()) e.food_name = "กรุณากรอกชื่ออาหาร";
    if (form.calories === "" || Number(form.calories) < 0)
      e.calories = "กรอกแคลอรี >= 0";
    if (form.price === "" || Number(form.price) < 0) e.price = "กรอกราคา >= 0";
    if (!form.entry_date) e.entry_date = "กรุณาเลือกวันที่";
    if (isAdmin && mode === "create" && !form.user_id)
      e.user_id = "กรุณาเลือกเจ้าของ";
    return e;
  }

  async function handleSubmit() {
    const e = validate();
    if (Object.keys(e).length) {
      setErrors(e);
      return;
    }

    setLoading(true);
    setSubmitError('');
    try {
      const payload: FoodEntryPayload = {
        food_name: form.food_name.trim(),
        calories: Number(form.calories),
        price: Number(form.price),
        entry_date: `${form.entry_date}T12:00:00Z`,
      };
      if (isAdmin && mode === "create") payload.user_id = Number(form.user_id);
      await onSubmit(payload);
      onClose();
    } catch (e) {
      setSubmitError((e as Error).message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
      <DialogTitle>
        {mode === "create" ? "เพิ่มรายการอาหาร" : "แก้ไขรายการอาหาร"}
      </DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ mt: 1 }}>
          {isAdmin && mode === "create" && (
            <TextField
              select
              label="เจ้าของ"
              value={form.user_id}
              onChange={(e) => set("user_id", e.target.value)}
              error={!!errors.user_id}
              helperText={errors.user_id}
            >
              {OWNER_OPTIONS.map((o) => (
                <MenuItem key={o.value} value={o.value}>
                  {o.label}
                </MenuItem>
              ))}
            </TextField>
          )}
          <TextField
            label="ชื่ออาหาร"
            value={form.food_name}
            onChange={(e) => set("food_name", e.target.value)}
            error={!!errors.food_name}
            helperText={errors.food_name}
          />
          <TextField
            label="แคลอรี (kcal)"
            type="number"
            value={form.calories}
            onChange={(e) => set("calories", e.target.value)}
            error={!!errors.calories}
            helperText={errors.calories}
            slotProps={{ htmlInput: { min: 0 } }}
          />
          <TextField
            label="ราคา (฿)"
            type="number"
            value={form.price}
            onChange={(e) => set("price", e.target.value)}
            error={!!errors.price}
            helperText={errors.price}
            slotProps={{ htmlInput: { min: 0, step: 0.01 } }}
          />
          <TextField
            label="วันที่"
            type="date"
            value={form.entry_date}
            onChange={(e) => set("entry_date", e.target.value)}
            error={!!errors.entry_date}
            helperText={errors.entry_date}
            slotProps={{ inputLabel: { shrink: true } }}
          />
        </Stack>
      </DialogContent>
      <DialogActions sx={{ flexDirection: 'column', alignItems: 'stretch', gap: 1, px: 3, pb: 2 }}>
        {submitError && (
          <Typography variant="body2" color="error">{submitError}</Typography>
        )}
        <Stack direction="row" sx={{ justifyContent: 'flex-end' }} spacing={1}>
          <Button onClick={onClose} disabled={loading}>ยกเลิก</Button>
          <Button variant="contained" onClick={handleSubmit} disabled={loading}>
            {loading ? "กำลังบันทึก..." : "บันทึก"}
          </Button>
        </Stack>
      </DialogActions>
    </Dialog>
  );
}
