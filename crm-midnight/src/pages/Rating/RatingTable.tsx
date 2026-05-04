import { Paper, TableContainer, TableHead, TableRow, TableBody, TableCell, Stack, Avatar } from '@mui/material'
import { Table, Box } from '@mui/material' 
import { StyledBox, StyledTableCell, StyledTypography } from './styles'
import FlashOnIcon from "@mui/icons-material/FlashOn"; 
import { RatingType } from '../../types';

interface RatingTableProps {
  rows: RatingType[];
  currentUserId?: string;
}

const getDisplayName = (row: RatingType) =>
  row.user.nick_name || row.user.first_name || row.user.username || row.user.user_id;

export default function RatingTable({ rows, currentUserId }: RatingTableProps) {
  return (
    <StyledBox>
      <Paper
        sx={{
          width: "100%",
          maxWidth: 600,
          borderRadius: "12px",
          boxShadow: 'none',
          backgroundColor: "#151A22",
          height: "auto",
        }}
      >
        <TableContainer>
          <Table aria-label="leaderboard table" sx={{borderCollapse: 'separate', borderSpacing: '0 4px'}}>
            <TableHead>
              <TableRow>
                <StyledTableCell>НИКНЕЙМ</StyledTableCell>
                <StyledTableCell align="right">НОКАУТЫ</StyledTableCell>
                <StyledTableCell align="right">РЕЙТИНГ</StyledTableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {rows?.map((row) => (
                <TableRow
                  key={row.user.user_id}
                  sx={{
                    height: "28px",
                    "& td, & th": {
                      height: 28,
                      py: 0,
                      borderBottom: 'none',
                    },
                    background:
                      row.rank === 1
                      ? '#F7E74D1F'
                      : row.rank === 2
                      ? '#F6F6F61F'
                      : row.rank === 3
                      ? '#F2CBA91F'
                      : 'transparent',
                    ...(row.user.user_id === currentUserId && {
                      outline: "1px solid #B9B9B9",
                      borderRadius: "8px",
                    }),
                  }}
                >
                  <TableCell component="th" scope="row" sx={{ padding: 0 }}>
                    <Stack direction="row" alignItems="center" spacing={2}>
                      <Box
                        sx={{
                          width: 20,
                          height: 20,
                          borderRadius: "50%",
                          display: "flex",
                          justifyContent: "center",
                          alignItems: "center",
                          color: "white",
                          background:
                            row.rank === 1
                              ? "linear-gradient(to right, #F8E84E, #E4D43A)"
                              : row.rank === 2
                              ? "linear-gradient(to right, #F6F6F6, #949494)"
                              : row.rank === 3
                              ? "linear-gradient(to right, #F2CBA9, #D88B53)"
                              : "transparent",
                          fontSize: "13px",
                          fontWeight: "500",
                        }}
                      >
                        {row.rank > 3 ? (
                          <StyledTypography variant="body1" sx={{ color: "#fff!important" }}>
                            {row.rank}
                          </StyledTypography>
                        ) : (
                          <StyledTypography variant="body1">
                            {row.rank}
                          </StyledTypography>
                        )}
                      </Box>
                      <Avatar
                        src={row.user.photo_url}
                        alt="avatar"
                        sx={{ width: 22, height: 22 }}
                      />
                      <StyledTypography variant="body1" sx={{ color: "#fff!important", fontSize: '15px' }}>
                        {getDisplayName(row)}
                      </StyledTypography>
                    </Stack>
                  </TableCell>
                  <TableCell align="right" padding="none">
                    <StyledTypography variant="body1" sx={{ color: "#fff!important", fontSize: '15px' }}>
                      {/* {row.knockouts} - добавьте поле knockouts в тип RatingType если нужно */}
                      —
                    </StyledTypography>
                  </TableCell>
                  <TableCell align="right">
                    <Stack direction="row" alignItems="center" justifyContent="flex-end" spacing={0.5}>
                      <StyledTypography variant="body1" sx={{ color: "#fff!important", fontSize: '15px' }}>
                        {row.points}
                      </StyledTypography>
                      <FlashOnIcon sx={{ color: "gold", fontSize: "1rem" }} />
                    </Stack>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>
    </StyledBox>
  );
}
