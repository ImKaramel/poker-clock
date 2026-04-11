import React, { useState } from "react";
import { profileAPI } from "../../../utils/api";
import { Title, Subtitle, Container, Input, Button } from "./styles";
import { useNavigate } from "react-router-dom";

export default function Welcome() {
  const [nickname, setNickname] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const handleSubmit = async () => {
    if (!nickname.trim()) return;

    try {
      setLoading(true);

      await profileAPI.updateProfile(nickname);

      navigate('/useragreement')

    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container>
      <Title>ДОБРО ПОЖАЛОВАТЬ в  Midnight Club</Title>
      <Subtitle>Как к вам обращаться?</Subtitle>

      <Input
        placeholder="Введите ник"
        value={nickname}
        onChange={(e) => setNickname(e.target.value)}
      />

      <Button onClick={handleSubmit} disabled={loading}>
        {loading ? "Сохранение..." : "Принять"}
      </Button>
    </Container>
  );
}