import { Grid, Card, CardContent, Typography, Chip, Box, LinearProgress } from "@mui/material";
import TrendingUpIcon from "@mui/icons-material/TrendingUp";
import AttachMoneyIcon from "@mui/icons-material/AttachMoney";
import type { DailySummary } from "../types/api";

interface StatCardProps {
  title: string;
  value: string | number;
  limit: string | number;
  exceeded: boolean;
  unit?: string;
  icon: React.ReactNode;
}

function StatCard({
  title,
  value,
  limit,
  exceeded,
  unit = "",
  icon,
}: StatCardProps) {
  const progress = Math.min((Number(value) / Number(limit)) * 100, 100);
  return (
    <Card sx={{ height: "100%" }}>
      <CardContent>
        <Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 1 }}>
          <Box sx={{ color: exceeded ? "error.main" : "primary.main" }}>
            {icon}
          </Box>
          <Typography variant="body2" sx={{ color: "text.secondary", fontWeight: 500 }}>
            {title}
          </Typography>
        </Box>
        <Typography
          variant="h4"
          sx={{ fontWeight: "bold", color: exceeded ? "error.main" : "text.primary" }}
        >
          {value}
          <Typography component="span" variant="body1" sx={{ color: "text.secondary" }}>
            {unit}
          </Typography>
        </Typography>
        <Typography variant="body2" sx={{ color: "text.secondary", mt: 0.5 }}>
          เป้าหมาย: {limit}{unit}
        </Typography>
        <LinearProgress
          variant="determinate"
          value={progress}
          color={exceeded ? "error" : "primary"}
          sx={{ mt: 1.5, borderRadius: 1 }}
        />
        {exceeded && (
          <Chip label="เกิน limit" color="error" size="small" sx={{ mt: 1 }} />
        )}
      </CardContent>
    </Card>
  );
}

interface Props {
  summary: DailySummary | null;
}

export default function SummaryCards({ summary }: Readonly<Props>) {
  if (!summary) return null;
  return (
    <>
      <Typography
        variant="subtitle2"
        sx={{ mb: 1.5, fontWeight: 600, color: "text.secondary", textTransform: "uppercase", letterSpacing: 0.5 }}
      >
        สรุปประจำวัน
      </Typography>
      <Grid container spacing={2} sx={{ mb: 4 }}>
        <Grid size={{ xs: 12, sm: 6 }}>
          <StatCard
            title="แคลอรีวันนี้"
            value={summary.total_calories}
            limit={summary.calorie_limit}
            exceeded={summary.calorie_exceeded}
            unit=" kcal"
            icon={<TrendingUpIcon />}
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6 }}>
          <StatCard
            title="ค่าใช้จ่ายเดือนนี้"
            value={Number(summary.total_price).toFixed(2)}
            limit={summary.price_limit}
            exceeded={summary.price_exceeded}
            unit=" ฿"
            icon={<AttachMoneyIcon />}
          />
        </Grid>
      </Grid>
    </>
  );
}
