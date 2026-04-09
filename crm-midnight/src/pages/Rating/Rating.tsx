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

export default function Rating() {
  const [series, setSeries] = React.useState("Мартовская серия");
  const [loading, setLoading] = useState(true);
  const [rating, setRating] = useState<RatingType[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const getRating = async () => {
      try {
        setLoading(true);
        const response = await ratingAPI.getRating();
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
  }, []);

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
              <MenuItem value="Мартовская серия">Мартовская серия</MenuItem>
              <MenuItem value="Апрельская серия">Апрельская серия</MenuItem>
              <MenuItem value="Майская серия">Майская серия</MenuItem>
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
