const express = require('express');
   const { Pool } = require('pg');
   const cors = require('cors');
   require('dotenv').config();

   const app = express();
   const pool = new Pool({ connectionString: process.env.DATABASE_URL });

   app.use(cors());
   app.use(express.json());
   app.use(express.static('.'));

   // Get all todos
   app.get('/todos', async (req, res) => {
     try {
       const result = await pool.query('SELECT * FROM todos ORDER BY id');
       res.json(result.rows);
     } catch (err) {
       res.status(500).json({ error: err.message });
     }
   });

   // Add a todo
   app.post('/todos', async (req, res) => {
     const { title } = req.body;
     try {
       const result = await pool.query(
         'INSERT INTO todos (title, completed) VALUES ($1, $2) RETURNING *',
         [title, false]
       );
       res.json(result.rows[0]);
     } catch (err) {
       res.status(500).json({ error: err.message });
     }
   });

   // Delete a todo
   app.delete('/todos/:id', async (req, res) => {
     const { id } = req.params;
     try {
       await pool.query('DELETE FROM todos WHERE id = $1', [id]);
       res.json({ success: true });
     } catch (err) {
       res.status(500).json({ error: err.message });
     }
   });

   app.listen(process.env.PORT, () => {
     console.log(`Server running on port ${process.env.PORT}`);
   });

   // Mark a todo as complete/incomplete
app.put('/todos/:id', async (req, res) => {
  const { id } = req.params;
  const { completed } = req.body;
  try {
    const result = await pool.query(
      'UPDATE todos SET completed = $1 WHERE id = $2 RETURNING *',
      [completed, id]
    );
    res.json(result.rows[0]);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});