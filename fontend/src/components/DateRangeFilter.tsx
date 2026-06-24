import { useState } from "react";
import { Stack, TextField, Button, Typography } from "@mui/material";
import FilterListIcon from "@mui/icons-material/FilterList";
import RestartAltIcon from "@mui/icons-material/RestartAlt";

function todayStr(): string {
  return new Date().toISOString().slice(0, 10);
}

interface Props {
  onApply: (from: string, to: string) => void;
}

export default function DateRangeFilter({ onApply }: Props) {
  const [from, setFrom] = useState(todayStr);
  const [to, setTo] = useState(todayStr);
  const [error, setError] = useState("");

  function apply() {
    if (from > to) {
      setError("วันที่เริ่มต้นต้องไม่เกินวันที่สิ้นสุด");
      return;
    }
    setError("");
    onApply(from, to);
  }

  function reset() {
    const t = todayStr();
    setFrom(t);
    setTo(t);
    setError("");
    onApply(t, t);
  }

  return (
    <Stack
      direction={{ xs: "column", sm: "row" }}
      spacing={1.5}
      sx={{ alignItems: 'flex-start', flexWrap: 'wrap' }}
    >
      <TextField
        label="วันที่เริ่มต้น"
        type="date"
        value={from}
        onChange={(e) => {
          setFrom(e.target.value);
          setError("");
        }}
        slotProps={{ inputLabel: { shrink: true } }}
        size="small"
        sx={{ minWidth: 160 }}
      />
      <TextField
        label="วันที่สิ้นสุด"
        type="date"
        value={to}
        onChange={(e) => {
          setTo(e.target.value);
          setError("");
        }}
        slotProps={{ inputLabel: { shrink: true } }}
        size="small"
        sx={{ minWidth: 160 }}
      />
      <Button
        variant="contained"
        startIcon={<FilterListIcon />}
        onClick={apply}
        sx={{ alignSelf: "center" }}
      >
        ค้นหา
      </Button>
      <Button
        variant="outlined"
        startIcon={<RestartAltIcon />}
        onClick={reset}
        sx={{ alignSelf: "center" }}
      >
        รีเซ็ต
      </Button>
      {error && (
        <Typography variant="body2" sx={{ color: 'error.main', alignSelf: "center" }}>
          {error}
        </Typography>
      )}
    </Stack>
  );
}
