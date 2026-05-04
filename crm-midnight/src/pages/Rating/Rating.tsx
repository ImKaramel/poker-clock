import React, { useEffect, useState } from "react";
import {
  BluredPoint,
  RatingContainer,
  RatingHeaderContainer,
  RatingHeaderWrapper,
  RatingPeriodContainer,
  RatingTitle,
  StyledSelect,
} from "./styles";
import MenuItem from "@mui/material/MenuItem";
import { Chip } from "@mui/material";
import RatingTable from "./RatingTable";
import { RatingType } from "../../types";
import { ratingAPI } from "../../utils/api";

const monthNames = [
  "Январская", "Февральская", "Мартовская", "Апрельская", "Майская", "Июньская",
  "Июльская", "Августовская", "Сентябрьская", "Октябрьская", "Ноябрьская", "Декабрьская",
];

const currentSeriesName = () => `${monthNames[new Date().getMonth()]} серия`;
const monthParamBySeries = (seriesName: string) => {
  const monthIndex = monthNames.findIndex((month) => seriesName.startsWith(month));
  const date = new Date();
  const selectedMonth = monthIndex >= 0 ? monthIndex : date.getMonth();
  return `${date.getFullYear()}-${String(selectedMonth + 1).padStart(2, "0")}`;
};

export default function Rating() {
  const [series, setSeries] = React.useState(currentSeriesName());
   // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [loading, setLoading] = useState(true);
  const [rating, setRating] = useState<RatingType[]>([]);
   // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const timer = window.setInterval(() => {
      setSeries(currentSeriesName());
    }, 60 * 60 * 1000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    const getRating = async () => {
      try {
        setLoading(true);
        const response = await ratingAPI.getRating(monthParamBySeries(series));
        setRating(response.data);
        setError(null);
      } catch (err: any) {
        console.error("Error fetching rating:", err);
        setError(err.message || "Ошибка загрузки рейтинга");
      } finally {
        setLoading(false);
      }
    };

    getRating();
  }, [series]);

  const handleChange = (event: any) => {
    setSeries(event.target.value);
  };
  return (
    <RatingContainer>
      <RatingHeaderContainer>
        <BluredPoint />
        <RatingHeaderWrapper>
          <RatingTitle>Рейтинг</RatingTitle>
          <RatingPeriodContainer>
            <div>
              <Chip
                label="Сезонный"
                sx={{
                  width: 126,
                  bgcolor: "white",
                  color: "black",
                }}
              />
              <Chip
                label="Глобальный"
                variant="outlined"
                sx={{
                  width: 126,
                  bgcolor: "#2C2C2E",
                  borderColor: "#A0A0A0",
                  color: "#A0A0A0",
                }}
              />
            </div>
            <StyledSelect
              value={series}
              onChange={handleChange}
              displayEmpty
              inputProps={{ "aria-label": "Выбор серии" }}
            >
              {monthNames.map((month) => {
                const value = `${month} серия`;
                return (
                  <MenuItem key={value} value={value}>
                    {value}
                  </MenuItem>
                );
              })}
            </StyledSelect>
          </RatingPeriodContainer>
        </RatingHeaderWrapper>
      </RatingHeaderContainer>
      <RatingHeaderContainer style={{ marginTop: "5px", height: "auto" }}>
        <RatingHeaderWrapper style={{ height: "auto" }}>
          <RatingTable rows={rating} />
        </RatingHeaderWrapper>
      </RatingHeaderContainer>
    </RatingContainer>
  );
}
