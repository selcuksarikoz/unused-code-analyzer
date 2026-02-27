// Complex React Component Test
// Testing class components and complex patterns

import React, { Component, PureComponent } from 'react';
import type { ReactNode, CSSProperties } from 'react';

// ========== CLASS COMPONENT ==========

interface ComplexComponentProps {
  title: string;
  data: unknown[];
  onUpdate: (data: unknown[]) => void;
  unusedOptional?: string;
}

interface ComplexComponentState {
  loading: boolean;
  error: Error | null;
  unusedField: string;
}

export class ComplexComponent extends Component<ComplexComponentProps, ComplexComponentState> {
  // Static properties
  static defaultProps = {
    unusedOptional: 'default'
  };
  
  // Used private field
  private timerId: number | null = null;
  
  // Unused private field
  private unusedPrivate: string = 'unused';
  
  constructor(props: ComplexComponentProps) {
    super(props);
    
    this.state = {
      loading: false,
      error: null,
      unusedField: 'never used'
    };
  }
  
  // Lifecycle methods
  componentDidMount() {
    this.loadData();
  }
  
  componentDidUpdate(prevProps: ComplexComponentProps) {
    if (prevProps.data !== this.props.data) {
      this.processData();
    }
  }
  
  componentWillUnmount() {
    if (this.timerId) {
      clearTimeout(this.timerId);
    }
  }
  
  // Unused lifecycle
  shouldComponentUpdate(): boolean {
    return true;
  }
  
  // Methods
  private loadData() {
    this.setState({ loading: true });
    // Load data logic
  }
  
  private processData() {
    // Process logic
    this.setState({ loading: false });
  }
  
  // Unused method
  private unusedMethod() {
    console.log('never called');
  }
  
  // Method with unused parameter
  private handleEvent(event: Event, unusedDetail: unknown) {
    console.log('event', event.type);
  }
  
  render() {
    const { title, data } = this.props;
    const { loading, error } = this.state;
    
    if (loading) return <div>Loading...</div>;
    if (error) return <div>Error: {error.message}</div>;
    
    return (
      <div>
        <h1>{title}</h1>
        <pre>{JSON.stringify(data, null, 2)}</pre>
      </div>
    );
  }
}

// ========== PURE COMPONENT ==========

interface ListItemProps {
  id: number;
  name: string;
  unusedDescription?: string;
}

export class ListItem extends PureComponent<ListItemProps> {
  // Unused static
  static unusedStatic = 'static';
  
  render() {
    const { id, name } = this.props;
    return <li>{id}: {name}</li>;
  }
}

// ========== HIGHER-ORDER COMPONENT ==========

// Unused HOC
function withUnusedData<P extends object>(
  WrappedComponent: React.ComponentType<P>
): React.ComponentType<P> {
  return class extends Component<P> {
    render() {
      return <WrappedComponent {...this.props} />;
    }
  };
}

// ========== STYLES ==========

// Unused styles
const unusedStyles: CSSProperties = {
  color: 'red',
  fontSize: '14px'
};

// Used styles
const containerStyles: CSSProperties = {
  padding: '20px'
};

// ========== UTILITY COMPONENTS ==========

// Unused utility component
const UnusedWrapper: React.FC<{ children: ReactNode }> = ({ children }) => {
  return <div className="wrapper">{children}</div>;
};
