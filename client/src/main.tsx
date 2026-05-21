import { createRoot } from 'react-dom/client';
import './app.css';
import App from './App';

const rootElement = document.getElementById('app');

if (!rootElement) {
  throw new Error('Root element #app not found');
}

createRoot(rootElement).render(<App />);
