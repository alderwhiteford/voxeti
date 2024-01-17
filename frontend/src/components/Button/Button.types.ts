import { ReactNode } from "react";

export interface StyledButtonProps {
  children: ReactNode;
  href?: string;
  type?: "submit" | "button";
  icon?: string;
  disabled?: boolean;
  size?: "sm" | "md" | "lg";
  color:
    | "primary"
    | "seconday"
    | "producer"
    | "designer"
    | "delete"
    | "success";
  onClick?: () => void;
}
