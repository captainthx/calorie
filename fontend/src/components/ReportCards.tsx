import { Grid, Card, CardContent, Typography, Box } from "@mui/material";
import BarChartIcon from "@mui/icons-material/BarChart";
import TrendingUpIcon from "@mui/icons-material/TrendingUp";
import TrendingDownIcon from "@mui/icons-material/TrendingDown";
import TrendingFlatIcon from "@mui/icons-material/TrendingFlat";
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

export default function ReportCards({ report }: Readonly<Props>) {
  if (!report) return null;
  const diff = report.entries_comparison?.difference ?? 0;
  const diffLabel = diff > 0 ? `+${diff}` : String(diff);
  const diffColor = diff > 0 ? "success.main" : diff < 0 ? "error.main" : "text.secondary";
  const diffIcon = diff > 0 ? <TrendingUpIcon /> : diff < 0 ? <TrendingDownIcon /> : <TrendingFlatIcon />;
  const diffSubtitle = diff > 0 ? "มากกว่าสัปดาห์ก่อน" : diff < 0 ? "น้อยกว่าสัปดาห์ก่อน" : "เท่าสัปดาห์ก่อน";

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
                <Box sx={{ color: diffColor }}>{diffIcon}</Box>
                <Typography variant="body2" sx={{ color: "text.secondary", fontWeight: 500 }}>
                  เทียบสัปดาห์ก่อน
                </Typography>
              </Box>
              <Typography variant="h4" sx={{ fontWeight: "bold", color: diffColor }}>
                {diffLabel}
              </Typography>
              <Typography variant="body2" sx={{ color: "text.secondary", mt: 0.5 }}>
                {diffSubtitle}
              </Typography>
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
