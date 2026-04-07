import styled from "styled-components";
import { Box, Select, TableCell, Typography } from '@mui/material'

export const RatingContainer = styled.div`
    width: 100%;
    border-radius: 25px;
    background-color: #0B0B0B;
`
export const RatingHeaderContainer = styled.div`
    position: relative;
    width: 100%;
    height: 174px;
    border-radius: 25px;
    background-color: #151A22;
    overflow: hidden;
`
export const RatingHeaderWrapper = styled.div`
    width: auto;
    height: 124px;
    padding: 34px 20px 20px 20px;
`
export const RatingTitle = styled.div`
    width: auto;
    height: 24px;
    font-weight: 500;
    font-size: 24px;
    line-height: 83%;
    color: #FFFFFF;
`
export const BluredPoint = styled.div`
    position: absolute;
    top: -114px;
    left: 205px;
    width: 289px;
    height: 168px;
    background-color: #FFFFFF4D;
    filter: blur(60.7px);
`
export const RatingPeriodContainer = styled.div`
    width: auto;
    height: 80px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    margin-top: 20px;
`
export const StyledSelect = styled(Select)({
    width: 388,
    height: 38,
    borderRadius: '25px!important',
    border: '1px solid #7E7E7E',
    '& .MuiOutlinedInput-root': {
      height: '100%',
      borderRadius: '25px',
      backgroundColor: '#1A1A1A',
      color: '#FFFFFF',
      '& fieldset': {
        borderColor: 'transparent',
      },
      '&:hover fieldset': {
        borderColor: 'transparent',
      },
      '&.Mui-focused fieldset': {
        borderColor: 'transparent',
      },
    },
    '& .MuiSelect-select': {
      height: '100% !important',
      paddingTop: 0,
      paddingBottom: 0,
      display: 'flex',
      alignItems: 'center',
      boxSizing: 'border-box', 
      color: '#fff'
    },
    '& .MuiSelect-icon': {
      color: '#FFFFFF',
    },
  })
  
  export const StyledBox = styled(Box)({
    display: 'flex',
    width: '100%',
    height: 'auto',
    backgroundColor: 'transparent'
  })
export const StyledTableCell = styled(TableCell)({
  padding: '0!important',
  color: '#FFFFFF66!important',
  fontWeight: '500!important',
  fontSize: '10px!important',
  lineHeight: '83%!important',
  paddingBottom: '11px!important'
})

export const StyledTypography = styled(Typography)({
  fontWeight: '500!important',
  fontSize: '13px!important',
  lineHeight: '83%!important',
  color: '#151A22!important',
  fontFamily: 'inherit!important'
})