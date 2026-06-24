import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  IconButton,
  Tooltip,
  Typography,
  Box,
} from "@mui/material";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/Delete";
import type { FoodEntry } from "../types/api";

interface Props {
  entries: FoodEntry[];
  isAdmin?: boolean;
  onEdit: (entry: FoodEntry) => void;
  onDelete: (entry: FoodEntry) => void;
}

export default function FoodTable({
  entries = [],
  isAdmin = false,
  onEdit,
  onDelete,
}: Readonly<Props>) {
  if (!entries.length) {
    return (
      <Box sx={{ py: 6, textAlign: 'center' }}>
        <Typography sx={{ color: 'text.secondary' }}>
          ไม่มีรายการในช่วงเวลาที่เลือก
        </Typography>
      </Box>
    );
  }

  return (
    <TableContainer component={Paper} elevation={1}>
      <Table size="small">
        <TableHead>
          <TableRow sx={{ bgcolor: "primary.main" }}>
            {isAdmin && (
              <TableCell sx={{ color: "white", fontWeight: 600 }}>
                ชื่อผู้ใช้
              </TableCell>
            )}
            <TableCell sx={{ color: "white", fontWeight: 600 }}>
              ชื่ออาหาร
            </TableCell>
            <TableCell align="right" sx={{ color: "white", fontWeight: 600 }}>
              แคลอรี (kcal)
            </TableCell>
            <TableCell align="right" sx={{ color: "white", fontWeight: 600 }}>
              ราคา (฿)
            </TableCell>
            <TableCell sx={{ color: "white", fontWeight: 600 }}>
              วันที่บันทึก
            </TableCell>
            <TableCell align="center" sx={{ color: "white", fontWeight: 600 }}>
              จัดการ
            </TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {entries.map((row) => (
            <TableRow
              key={row.id}
              hover
              sx={{ "&:last-child td": { border: 0 } }}
            >
              {isAdmin && (
                <TableCell>{row.user_name ?? `User ${row.user_id}`}</TableCell>
              )}
              <TableCell>{row.food_name}</TableCell>
              <TableCell align="right">{row.calories}</TableCell>
              <TableCell align="right">
                {Number(row.price).toFixed(2)}
              </TableCell>
              <TableCell>
                {row.entry_date ? row.entry_date.slice(0, 10) : "-"}
              </TableCell>
              <TableCell align="center">
                <Tooltip title="แก้ไข">
                  <IconButton
                    size="small"
                    color="primary"
                    onClick={() => onEdit(row)}
                  >
                    <EditIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
                <Tooltip title="ลบ">
                  <IconButton
                    size="small"
                    color="error"
                    onClick={() => onDelete(row)}
                  >
                    <DeleteIcon fontSize="small" />
                  </IconButton>
                </Tooltip>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  );
}
