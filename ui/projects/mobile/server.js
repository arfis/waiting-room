const express = require('express');
const path = require('path');
const app = express();
const port = 4204;

// Serve static files from the dist directory
app.use(express.static(path.join(__dirname, '../../dist/mobile/browser')));

// Handle Angular routing - return index.html for all routes
app.get('*', (req, res) => {
  res.sendFile(path.join(__dirname, '../../dist/mobile/browser/index.html'));
});

app.listen(port, () => {
  console.log(`Mobile app server running on http://localhost:${port}`);
});
