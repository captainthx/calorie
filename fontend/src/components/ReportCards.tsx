import { Grid, Card, CardContent, Typography, Box, Chip } from "@mui/material";
import BarChartIcon from "@mui/icons-material/BarChart";
import CompareArrowsIcon from "@mui/icons-material/CompareArrows";
import LocalFireDepartmentIcon from "@mui/icons-material/LocalFireDepartment";
import type { AdminReport } from "../types/api";

interface StatCardProps {
  title: string;
  value: string | number | null | undefined;
  icon: React.ReactNode;
  highlight?: boolean;
}

function StatCard({ title, value, icon, highlight = false }: StatCardProps) {
  return (
    <Card sx={{ height: "100%" }}>
      <CardContent>
        <Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 1 }}>
          <Box sx={{ color: "secondary.main" }}>{icon}</Box>
          <Typography variant="body2" sx={{ color: "text.secondary", fontWeight: 500 }}>
            {title}
          </Typography>
        </Box>
        <Typography
          variant="h4"
          sx={{ fontWeight: "bold", color: highlight ? "secondary.main" : "text.primary" }}
        >
          {value ?? "-"}
        </Typography>
      </CardContent>
    </Card>
  );
}

interface Props {
  report: AdminReport | null;
}

export default function ReportCards({ report }: Props) {
  if (!report) return null;
  const diff = report.entries_comparison?.difference ?? 0;
  const diffLabel = diff > 0 ? `+${diff}` : String(diff);

  return (
    <>
      <Typography
        variant="subtitle2"
        sx={{ mb: 1.5, fontWeight: 600, color: "text.secondary", textTransform: "uppercase", letterSpacing: 0.5 }}
      >
        รายงานภาพรวม
      </Typography>
      <Grid container spacing={2} sx={{ mb: 4 }}>
        <Grid size={{ xs: 6, sm: 3 }}>
          <StatCard
            title="รายการ 7 วันล่าสุด"
            value={report.entries_last_7_days}
            icon={<BarChartIcon />}
          />
        </Grid>
        <Grid size={{ xs: 6, sm: 3 }}>
          <StatCard
            title="รายการ 7 วันก่อน"
            value={report.entries_previous_7_days}
            icon={<BarChartIcon />}
          />
        </Grid>
        <Grid size={{ xs: 6, sm: 3 }}>
          <Card sx={{ height: "100%" }}>
            <CardContent>
              <Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 1 }}>
                <Box sx={{ color: "secondary.main" }}>
                  <CompareArrowsIcon />
                </Box>
                <Typography variant="body2" sx={{ color: "text.secondary", fontWeight: 500 }}>
                  ผลต่าง
                </Typography>
              </Box>
              <Chip
                label={diffLabel}
                color={diff > 0 ? "success" : diff < 0 ? "error" : "default"}
                size="medium"
                sx={{ fontSize: "1.1rem", height: 36, px: 1 }}
              />
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 6, sm: 3 }}>
          <StatCard
            title="เฉลี่ย kcal/user"
            value={report.average_calories_per_user_last_7_days?.toFixed(0)}
            icon={<LocalFireDepartmentIcon />}
            highlight
          />
        </Grid>
      </Grid>
    </>
  );
}
