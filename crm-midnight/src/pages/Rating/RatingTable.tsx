import { Paper, TableContainer, TableHead, TableRow, TableBody, TableCell, Stack, Avatar } from '@mui/material'
import { Table, Box } from '@mui/material' 
import { StyledBox, StyledTableCell, StyledTypography } from './styles'
import FlashOnIcon from "@mui/icons-material/FlashOn";
import { faker } from "@faker-js/faker";

interface LeaderboardData {
  id: string;
  rank: number;
  nickname: string;
  avatar: string;
  knockouts: number;
  rating: number;
}

interface Props {
    rows: LeaderboardData[];
    currentUserId?: string;
  }

export const generateLeaderboardData = (count: number): LeaderboardData[] => {
  return Array.from({ length: count }, (_, index) => ({
    id: faker.string.uuid(),
    rank: index + 1,
    nickname: faker.person.firstName(),
    avatar: faker.image.avatar(), // Using faker's built-in avatar which uses lorempixel.com
    knockouts: faker.number.int({ min: 1, max: 1000 }),
    rating: faker.number.int({ min: 300, max: 2000 }),
  }));
};

export const rows: LeaderboardData[] = generateLeaderboardData(20);

export default function RatingTable({ rows, currentUserId }: Props) {
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
                    {rows.map((row) => (
                      <TableRow
                        key={row.id}
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
                          ...(row.id === currentUserId && {
                            outline: "1px solid #B9B9B9",
                          }),
                        }}
                      >
                        <TableCell
                          component="th"
                          scope="row"
                          sx={{ padding: 0 }}
                        >
                          <Stack
                            direction="row"
                            alignItems="center"
                            spacing={2}
                          >
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
                                <StyledTypography
                                  variant="body1"
                                  sx={{ color: "#fff!important" }}
                                >
                                  {row.rank}
                                </StyledTypography>
                              ) : (
                                <StyledTypography
                                  variant="body1"
                                >
                                  {row.rank}
                                </StyledTypography>
                              )}
                            </Box>
                            {row.rank <= 3 && (
                              <Avatar
                                src={row.avatar}
                                alt={row.nickname}
                                sx={{ width: 22, height: 22 }}
                              />
                            )}
                            {row.rank > 3 && (
                              <Avatar
                                src={row.avatar}
                                alt={row.nickname}
                                sx={{ width: 22, height: 22 }}
                              />
                            )}
                            <StyledTypography variant="body1" sx={{ color: "#fff!important", fontSize: '15px' }}>
                              {row.nickname}
                            </StyledTypography>
                          </Stack>
                        </TableCell>
                        <TableCell align="right" padding="none">
                          <StyledTypography variant="body1" sx={{ color: "#fff!important", fontSize: '15px' }}>
                            {row.knockouts}
                          </StyledTypography>
                        </TableCell>
                        <TableCell align="right">
                          <Stack
                            direction="row"
                            alignItems="center"
                            justifyContent="flex-end"
                            spacing={0.5}
                          >
                            <StyledTypography variant="body1" sx={{ color: "#fff!important", fontSize: '15px' }}>
                              {row.rating}
                            </StyledTypography>
                            <FlashOnIcon
                              sx={{ color: "gold", fontSize: "1rem" }}
                            />
                          </Stack>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </Paper>
          </StyledBox>
  )
}
