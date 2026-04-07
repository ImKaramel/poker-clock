import { MenuContainer, MenuItem } from "./styles";
import { ReactComponent as Main } from "../../assets/main.svg";
import { ReactComponent as Tournament } from "../../assets/tournaments.svg";
import { ReactComponent as Rating } from "../../assets/rating.svg";
import { ReactComponent as Profile } from "../../assets/profile.svg";
import { useLocation, useNavigate } from "react-router-dom";

const Menu: React.FC<{ children?: React.ReactNode }> = ({ children }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const navItems = [
    { path: "/", image: <Main /> },
    { path: "/tournaments", image: <Tournament /> },
    { path: "/rating", image: <Rating /> },
    { path: "/profile", image: <Profile /> }
  ];
  return (
    <MenuContainer>
      {navItems.map((item) => {
        const isActive = location.pathname === item.path;

        return (
          <MenuItem
            key={item.path}
            $active={isActive}
            onClick={() => navigate(item.path)}
          >
            {item.image}
          </MenuItem>
        );
      })}
    </MenuContainer>
  );
};

export default Menu;
