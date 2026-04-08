import http from 'k6/http';
import { check, sleep } from 'k6';

// 1. Konfigurasi Load Test
export let options = {
    stages: [
        { duration: '10s', target: 100 },  // Naik perlahan ke 100 Virtual Users (VU)
        { duration: '30s', target: 1000 }, // Tahan di 1000 VU selama 30 detik
        { duration: '10s', target: 0 },    // Turun perlahan ke 0 VU
    ],
};

const BASE_URL = 'http://localhost:3000/v1'; 
const TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMiIsImVtYWlsIjoiaGVsbWlAZ21haWwuY29tIiwidXNlcm5hbWUiOiJIZWxtaSIsInJvbGUiOiJzdXBlcl9hZG1pbiIsImV4cCI6MTc3NTY1NDM1N30.ZTeqcsC7iHc-5vhBEBL7p0iVC5BLXJiTmm06cosSQks';   

export default function () {
    const headers = {
        'Authorization': `Bearer ${TOKEN}`,
        'Content-Type': 'application/json',
    };

    // ==========================================
    // GET - Read All Article Categories Doang
    // ==========================================
    let resGetAll = http.get(`${BASE_URL}/categories/article`, { headers });
    
    check(resGetAll, { 
        'GET /categories/article status 200': (r) => r.status === 200 
    });

    // Jeda 1 detik antar iterasi
    sleep(1);
}