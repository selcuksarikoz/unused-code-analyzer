// React Hooks Test File
// Testing unused hooks and hook-related code

import React, { 
  useState, 
  useEffect, 
  useCallback, 
  useMemo, 
  useRef,
  useContext,
  useReducer
} from 'react';

// Unused hook imports
import { 
  useLayoutEffect, 
  useImperativeHandle,
  useDebugValue 
} from 'react';

// ========== CUSTOM HOOKS ==========

// Used custom hook
function useCounter(initialValue: number = 0) {
  const [count, setCount] = useState(initialValue);
  
  const increment = useCallback(() => {
    setCount(c => c + 1);
  }, []);
  
  const decrement = useCallback(() => {
    setCount(c => c - 1);
  }, []);
  
  return { count, increment, decrement };
}

// Unused custom hook
function useUnusedCounter() {
  const [count, setCount] = useState(0);
  return { count, setCount };
}

// Custom hook with unused parameters
function useDataFetcher(url: string, unusedOptions: object) {
  const [data, setData] = useState(null);
  
  useEffect(() => {
    fetch(url)
      .then(r => r.json())
      .then(setData);
  }, [url]);
  
  return data;
}

// ========== CONTEXT ==========

// Used context
const ThemeContext = React.createContext('light');

// Unused context
const UnusedContext = React.createContext(null);

// ========== COMPONENT ==========

export function HooksDemo() {
  // Used hooks
  const [count, setCount] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  
  // Unused hooks results
  const [unusedState, setUnusedState] = useState('unused');
  const unusedRef = useRef(null);
  const unusedCallback = useCallback(() => {}, []);
  const unusedMemo = useMemo(() => count * 2, [count]);
  
  // Use the counter hook
  const { count: customCount, increment } = useCounter(10);
  
  // Used effect
  useEffect(() => {
    console.log('Count changed:', count);
  }, [count]);
  
  // Unused effect
  useEffect(() => {
    console.log('This effect is not properly configured');
  }, []);
  
  // Use context
  const theme = useContext(ThemeContext);
  
  // Unused context
  const unusedCtx = useContext(UnusedContext);
  
  // Event handler with unused parameter
  const handleClick = (event: React.MouseEvent, unusedParam: string) => {
    setCount(c => c + 1);
  };
  
  // Unused handler
  const unusedHandler = () => {
    console.log('never used');
  };
  
  return (
    <div>
      <h1>Hooks Demo</h1>
      <p>Count: {count}</p>
      <p>Custom Count: {customCount}</p>
      <button onClick={handleClick}>Increment</button>
      <button onClick={increment}>Custom Increment</button>
      <input ref={inputRef} />
      <p>Theme: {theme}</p>
    </div>
  );
}

// ========== REDUCER ==========

// Unused reducer
function unusedReducer(state: any, action: any) {
  switch (action.type) {
    case 'increment':
      return { count: state.count + 1 };
    default:
      return state;
  }
}

// Component with unused reducer
export function UnusedReducerComponent() {
  const [state, dispatch] = useReducer(unusedReducer, { count: 0 });
  
  // dispatch is used but state is not
  return <button onClick={() => dispatch({ type: 'increment' })}>+</button>;
}
