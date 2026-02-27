// React TypeScript component test file
// Testing TSX with React patterns

import React, { useState, useEffect, useCallback } from "react";
import type { FC, ReactNode } from "react";

// Unused imports
import { unusedHelper } from "./utils";
import type { UnusedProps } from "./utils";

// Import used
import { formatData } from "./utils";

// Type-only import used
import type { ComponentProps } from "./utils";

// ========== INTERFACES & TYPES ==========

interface ButtonProps {
  label: string;
  onClick: () => void;
  unusedColor?: string;
}

type UnusedCallback = () => void;

// ========== COMPONENTS ==========

// Component with unused props destructuring
export const Button: FC<ButtonProps> = ({ label, onClick, unusedColor }) => {
  const [count, setCount] = useState(0);
  const unusedState = "not used";
  
  // Unused handler
  const unusedHandler = () => {
    console.log("never used");
  };
  
  const handleClick = () => {
    setCount(c => c + 1);
    onClick();
  };
  
  return (
    <button onClick={handleClick}>
      {label} - {count}
    </button>
  );
};

// Component with unused parameters in callback
export const List: FC<{ items: string[] }> = ({ items }) => {
  const [unusedFlag, setUnusedFlag] = useState(false);
  
  const handleItemClick = useCallback((item: string, unusedIndex: number) => {
    console.log(item);
  }, []);
  
  return (
    <ul>
      {items.map((item, index) => (
        <li key={index} onClick={() => handleItemClick(item, index)}>
          {item}
        </li>
      ))}
    </ul>
  );
};

// Unused component
const UnusedComponent: FC = () => {
  return <div>Never rendered</div>;
};

// Component using type import
export const TypedComponent: FC<ComponentProps> = ({ title }) => {
  const data = formatData(title);
  return <h1>{data}</h1>;
};

// ========== HOOKS ==========

// Custom hook with unused parameter
function useCustomHook(unusedParam: string) {
  const [state, setState] = useState(0);
  
  useEffect(() => {
    console.log("effect");
  }, []);
  
  return state;
}

// ========== UTILITIES ==========

// Unused utility function
function unusedUtility(value: string): string {
  return value.trim();
}

// Utility with unused params
function processItems(items: string[], unusedFilter: string): string[] {
  return items.map(i => i.toUpperCase());
}

export default Button;
