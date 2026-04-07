import React from "react";
import styled from "styled-components";
import { useLocation, useNavigate } from "react-router-dom";
import { ReactComponent as CalendarIcon } from "../assets/calendar.svg";
import { ReactComponent as TrophyIcon } from "../assets/trophy.svg";
import { ReactComponent as UserIcon } from "../assets/user.svg";
import { ReactComponent as InfoIcon } from "../assets/info.svg";
import { ReactComponent as UserStarIcon } from "../assets/user-star.svg";

const ContentWrapper = styled.div`
  padding-bottom: 80px; 
`;
const Nav = styled.nav`
  display: flex;
  justify-content: space-around;
  background-color: #222;
  width: 90%;
  padding: 11px;
  position: fixed;
  bottom: 0;
  left: 50%;
  transform: translate(-50%, 0);
  
`;

const NavButton = styled.button<{ active: boolean }>`
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1;
  color: ${(props) =>
    props.active ? "rgb(254, 181, 0)" : "rgb(120, 120, 131)"};
  border: none;
  background-color: #222;
  cursor: pointer;
  transition: all 0.3s ease;

  :active{
    img {
      color: 'rgb(254, 181, 0)';
    }
  }
`;



const Layout: React.FC<{ children?: React.ReactNode }> = ({children}) => {
  const location = useLocation();
  const navigate = useNavigate();

  const navItems = [
    { path: "/", label: "Расписание", image: <CalendarIcon /> },
    { path: "/rating", label: "Рейтинг", image: <TrophyIcon /> },
    { path: "/profile", label: "Профиль", image: <UserIcon /> },
    { path: "/about", label: "О клубе", image: <InfoIcon /> },
    { path: "/admin", label: "Админ", image: <UserStarIcon /> },
  ];

  return (
    <>
      <ContentWrapper>{children}</ContentWrapper>
      <Nav>
        {navItems.map((item) => (
          <NavButton
            key={item.path}
            active={location.pathname === item.path}
            onClick={() => navigate(item.path)}
          >
            {item.image}
            {item.label}
          </NavButton>
        ))}
      </Nav>
      </>
  );
};

export default Layout;
