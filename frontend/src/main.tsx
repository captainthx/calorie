import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { ThemeProvider, createTheme, CssBaseline } from '@mui/material'
import App from './App'

const theme = createTheme({
  palette: {
    primary: { main: '#1565c0' },
    secondary: { main: '#6a1b9a' },
  },
  shape: { borderRadius: 10 },
  components: {
    MuiCard: { defaultProps: { elevation: 2 } },
    MuiButton: { defaultProps: { disableElevation: true } },
  },
})

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </ThemeProvider>
  </React.StrictMode>
)
