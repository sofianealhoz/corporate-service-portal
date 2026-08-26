import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import Catalogue from './Catalogue'
import ServiceDetail from './ServiceDetail'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Catalogue />} />
        <Route path="/services/:slug" element={<ServiceDetail />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
