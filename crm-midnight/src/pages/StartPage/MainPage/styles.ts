import styled from "styled-components";

export const Container = styled.div`
  box-sizing: border-box;
  height: 100vh;
  width: 100%;
  background: #000;
  color: white;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 20px;
`;

export const Title = styled.h1`
  font-size: 28px;
  font-weight: 700;
  text-align: center;
  margin-bottom: 10px;
`;

export const Subtitle = styled.p`
  text-align: center;
  opacity: 0.7;
  margin-bottom: 30px;
`;

export const Input = styled.input`
  width: calc(100% - 32px);
  padding: 16px;
  border-radius: 30px;
  border: 1px solid #333;
  background: transparent;
  color: white;
  outline: none;
  font-size: 16px;
  margin-bottom: 20px;
`;

export const Button = styled.button`
  width: 100%;
  padding: 16px;
  border-radius: 30px;
  border: none;
  background: white;
  color: black;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
`;
