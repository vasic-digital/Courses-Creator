import React from 'react';
import styled from 'styled-components';
import { Link, useLocation } from 'react-router-dom';
import { useMediaQuery } from '@/hooks';
import { theme } from '@/styles/theme';

const HeaderContainer = styled.header`
  background-color: ${props => props.theme.colors.background.secondary};
  border-bottom: 1px solid ${props => props.theme.colors.border};
  padding: 0 ${props => props.theme.spacing.lg};
  position: sticky;
  top: 0;
  z-index: ${props => props.theme.zIndex.dropdown};
  backdrop-filter: blur(8px);
`;

const HeaderContent = styled.div`
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
`;

const Logo = styled(Link)`
  display: flex;
  align-items: center;
  gap: ${props => props.theme.spacing.sm};
  font-size: ${props => props.theme.typography.fontSize.lg};
  font-weight: ${props => props.theme.typography.fontWeight.bold};
  color: ${props => props.theme.colors.text.primary};
  text-decoration: none;
  transition: color ${props => props.theme.transitions.fast};

  &:hover {
    color: ${props => props.theme.colors.primary};
  }

  svg {
    width: 32px;
    height: 32px;
    fill: ${props => props.theme.colors.primary};
  }
`;

const Navigation = styled.nav<{ $isMobile: boolean }>`
  display: ${props => props.$isMobile ? 'none' : 'flex'};
  align-items: center;
  gap: ${props => props.theme.spacing.lg};

  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    position: fixed;
    top: 64px;
    left: 0;
    right: 0;
    background: ${props => props.theme.colors.background.secondary};
    border-bottom: 1px solid ${props => props.theme.colors.border};
    padding: ${props => props.theme.spacing.lg};
    flex-direction: column;
    gap: ${props => props.theme.spacing.md};
    z-index: ${props => props.theme.zIndex.modal};
  }

  ${props => props.$isMobile && `
    display: flex;
  `}
`;

const NavLink = styled(Link)<{ $active: boolean }>`
  color: ${props => props.$active 
    ? props.theme.colors.primary 
    : props.theme.colors.text.secondary
  };
  text-decoration: none;
  font-weight: ${props => props.$active 
    ? props.theme.typography.fontWeight.semibold 
    : props.theme.typography.fontWeight.normal
  };
  padding: ${props => props.theme.spacing.sm} ${props => props.theme.spacing.md};
  border-radius: ${props => props.theme.borderRadius.md};
  transition: all ${props => props.theme.transitions.fast};

  &:hover {
    color: ${props => props.theme.colors.primary};
    background-color: ${props => props.theme.colors.background.tertiary};
  }
`;

const MobileMenuButton = styled.button`
  display: none;
  background: none;
  border: none;
  color: ${props => props.theme.colors.text.primary};
  cursor: pointer;
  padding: ${props => props.theme.spacing.sm};

  @media (max-width: ${props => props.theme.breakpoints.tablet}) {
    display: block;
  }
`;

const MobileMenuIcon = styled.div<{ $isOpen: boolean }>`
  width: 24px;
  height: 20px;
  position: relative;
  transform: ${props => props.$isOpen ? 'rotate(-45deg)' : 'none'};
  transition: transform ${props => props.theme.transitions.normal};

  span {
    position: absolute;
    height: 2px;
    width: 100%;
    background: ${props => props.theme.colors.text.primary};
    border-radius: 2px;
    transition: all ${props => props.theme.transitions.normal};
    left: 0;

    &:nth-child(1) {
      top: ${props => props.$isOpen ? '9px' : '0'};
      opacity: ${props => props.$isOpen ? '0' : '1'};
    }

    &:nth-child(2) {
      top: 9px;
    }

    &:nth-child(3) {
      top: ${props => props.$isOpen ? '9px' : '18px'};
      transform: ${props => props.$isOpen ? 'rotate(90deg)' : 'none'};
    }
  }
`;

const Header: React.FC = () => {
  const location = useLocation();
  const isMobile = useMediaQuery(`(max-width: ${theme.breakpoints.tablet})`);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = React.useState(false);

  const navigationItems = [
    { path: '/courses', label: 'Courses' },
    { path: '/about', label: 'About' },
    { path: '/help', label: 'Help' },
  ];

  const isActivePath = (path: string) => {
    if (path === '/courses') {
      return location.pathname.startsWith('/courses');
    }
    return location.pathname === path;
  };

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  return (
    <HeaderContainer>
      <HeaderContent>
        <Logo to="/">
          <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M4 6H20M4 12H20M4 18H20" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M9 6L7 8L9 10M15 6L17 8L15 10M9 18L7 20L9 22M15 18L17 20L15 22" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
          Course Player
        </Logo>

        {isMobile && (
          <MobileMenuButton onClick={toggleMobileMenu}>
            <MobileMenuIcon $isOpen={isMobileMenuOpen}>
              <span />
              <span />
              <span />
            </MobileMenuIcon>
          </MobileMenuButton>
        )}

        <Navigation $isMobile={isMobileMenuOpen}>
          {navigationItems.map(item => (
            <NavLink
              key={item.path}
              to={item.path}
              $active={isActivePath(item.path)}
              onClick={() => isMobile && setIsMobileMenuOpen(false)}
            >
              {item.label}
            </NavLink>
          ))}
        </Navigation>
      </HeaderContent>
    </HeaderContainer>
  );
};

export default Header;