import React, { useState } from "react";
import styled from "styled-components";

const ERROR_IMG_SRC =
  'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iODgiIGhlaWdodD0iODgiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgc3Ryb2tlPSIjMDAwIiBzdHJva2UtbGluZWpvaW49InJvdW5kIiBvcGFjaXR5PSIuMyIgZmlsbD0ibm9uZSIgc3Ryb2tlLXdpZHRoPSIzLjciPjxyZWN0IHg9IjE2IiB5PSIxNiIgd2lkdGg9IjU2IiBoZWlnaHQ9IjU2IiByeD0iNiIvPjxwYXRoIGQ9Im0xNiA1OCAxNi0xOCAzMiAzMiIvPjxjaXJjbGUgY3g9IjUzIiBjeT0iMzUiIHI9IjciLz48L3N2Zz4KCg==';

const Wrapper = styled.div<{ overlayColor: string; overlayOpacity: number }>`
  position: relative;
  width: 100%;
  height: 30%;
  overflow: hidden;
  
  &::after {
    content: "";
    position: absolute;
    inset: 0;
    background: ${(props) => props.overlayColor};
    opacity: ${(props) => props.overlayOpacity};
    pointer-events: none;
  }
`;

const StyledImg = styled.img`
  width: 100%;
  height: 100%;
  object-fit: cover;
`;

interface ImageWithFallbackProps
  extends React.ImgHTMLAttributes<HTMLImageElement> {
  overlayColor?: string;
  overlayOpacity?: number;
}

export const ImageWithFallback: React.FC<ImageWithFallbackProps> = ({
  src,
  alt,
  overlayColor = "black",
  overlayOpacity = 0.4,
  ...rest
}) => {
  const [didError, setDidError] = useState(false);

  const displayedSrc = didError ? ERROR_IMG_SRC : src;

  return (
    <Wrapper overlayColor={overlayColor} overlayOpacity={overlayOpacity}>
      <StyledImg
        src={displayedSrc}
        alt={alt}
        onError={() => setDidError(true)}
        {...rest}
      />
    </Wrapper>
  );
};
