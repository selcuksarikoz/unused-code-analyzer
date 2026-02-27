import React, { useState, useEffect } from 'react';
import { usedFunction, unusedExport } from './utils';

export function UsedComponent({ usedProp }) {
  const [state, setState] = useState(null);
  
  useEffect(() => {
    usedFunction();
    console.log('effect');
  }, []);
  
  return <div>{usedProp}</div>;
}

export const usedConst = "used";

function UnusedComponent() {
  return <span>unused</span>;
}

const localUnused = "unused";

export default function App() {
  return (
    <div>
      <UsedComponent usedProp="test" />
    </div>
  );
}
