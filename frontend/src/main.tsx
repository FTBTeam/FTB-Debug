import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'
import {Toaster} from "@/components/ui/toaster";

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <div className="h-screen w-full flex items-center justify-center">
            <App/>
            <Toaster/>
        </div>
    </React.StrictMode>
)
