import styled from "styled-components";

export const MenuContainer = styled.div`
    position: fixed;
    bottom: 24px;
    left: 50%;
    transform: translateX(-50%);
    width: 242px;
    height: 62px;
    border-radius: 81.32px;
    background-color: #FFFFFFB2;
    backdrop-filter: blur(20px);
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 2px;
    z-index: 1000;
`
export const MenuItem = styled.div<{ $active?: boolean }>`
    width: 58px;
    height: 58px;
    border-radius: 50%;
    border: 1px solid #FFFFFF;
    background-color: #FFFFFF9E;
    display: flex;
    align-items: center;
    justify-content: center;
    background: ${({ $active }) =>
    $active
      ? "#0F172A"
      : "rgba(255,255,255,0.6)"};

  svg {
    width: 24px;
    height: 24px;

    path {
      fill: ${({ $active }) => ($active ? "#fff" : "#1E1E1E")};
    }
  }
`