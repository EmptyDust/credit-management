import axios from 'axios';

const apiClient = axios.create({
  baseURL: '/api', // All requests will be sent to the api-gateway
  headers: {
    'Content-Type': 'application/json',
  },
});

// You can add interceptors for handling auth tokens here
// For example:
// apiClient.interceptors.request.use(config => {
//   const token = localStorage.getItem('token');
//   if (token) {
//     config.headers.Authorization = `Bearer ${token}`;
//   }
//   return config;
// });

export default apiClient; 