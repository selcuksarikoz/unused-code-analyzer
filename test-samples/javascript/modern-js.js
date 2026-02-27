// Modern ES6+ JavaScript Test File
// Testing various unused patterns in plain JS

// ========== IMPORTS ==========
// Used imports
import { fetchData, processData } from './api';

// Unused imports
import { unusedHelper, anotherUnused } from './api';
import DefaultExport from './api';

// ========== CONSTANTS ==========
// Used constant
const API_URL = 'https://api.example.com';

// Unused constants
const UNUSED_KEY = 'secret-key';
const UNUSED_CONFIG = { timeout: 5000 };

// ========== VARIABLES ==========
// Used variables
let counter = 0;
var message = 'Hello';

// Unused variables
let unusedCounter = 0;
var unusedMessage = 'Unused';
const unusedItems = [];

// ========== FUNCTIONS ==========
// Used function
export async function loadData() {
  const data = await fetchData(API_URL);
  return processData(data);
}

// Unused function
function unusedLoader() {
  return fetchData('/unused');
}

// Function with unused parameters
function handleSubmit(event, unusedFormData, unusedOptions) {
  event.preventDefault();
  console.log('Submitted');
}

// Arrow function with unused
const calculateTotal = (price, unusedTax) => {
  return price * 2;
};

// ========== CLASSES ==========
// Used class
export class DataManager {
  constructor() {
    this.data = [];
  }
  
  add(item) {
    this.data.push(item);
  }
}

// Unused class
class UnusedManager {
  constructor() {
    this.items = [];
  }
}

// ========== OBJECTS ==========
// Used object
const config = {
  apiUrl: API_URL,
  timeout: 5000
};

// Unused object
const unusedConfig = {
  debug: true,
  verbose: false
};

// ========== ARRAY DESTRUCTURING ==========
const [first, second] = [1, 2, 3];
const [usedItem, unusedItem] = ['a', 'b'];

// ========== OBJECT DESTRUCTURING ==========
const { name, unusedProp } = { name: 'test', unusedProp: 'value' };
const { usedField, anotherUnusedField } = { usedField: 1, anotherUnusedField: 2 };

// ========== ASYNC/AWAIT ==========
// Used async function
export async function fetchUsers() {
  const response = await fetch('/users');
  return response.json();
}

// Unused async function
async function unusedFetch() {
  return fetch('/unused');
}

// ========== PROMISES ==========
// Used promise
const dataPromise = fetchData('/data');

// Unused promise
const unusedPromise = fetch('/unused');

// ========== EVENT HANDLERS ==========
// Used handler
export function onClick(event) {
  console.log('Clicked', event.target);
}

// Unused handler
function onHover(unusedEvent) {
  console.log('Hovered');
}

// Use some variables to avoid all being flagged
console.log(counter, message, first, usedItem, name, usedField);
loadData();
