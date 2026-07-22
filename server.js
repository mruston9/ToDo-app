const express = require('express');
const { Pool } = require('pg');
const cors = require('cors');
require('dotenv').config();

const app = express();
const pool = new Pool({ connectionString: process.env.DATABASE_URL });

app.use(cors());
app.use(express.json());
app.use(express.static('.'));

// Get all todos (ordered by position)
app.get('/todos', async (req, res) => {
  try {
    const result = await pool.query('SELECT * FROM todos ORDER BY position');
    res.json(result.rows);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// Add a todo (placed at the bottom)
app.post('/todos', async (req, res) => {
  const { title } = req.body;
  try {
    const result = await pool.query(
      'INSERT INTO todos (title, completed, position) VALUES ($1, $2, COALESCE((SELECT MAX(position) FROM todos), 0) + 1) RETURNING *',
      [title, false]
    );
    res.json(result.rows[0]);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// Move a todo up or down
app.post('/todos/:id/move', async (req, res) => {
  const { id } = req.params;
  const { direction } = req.body;
  try {
    const current = await pool.query('SELECT position FROM todos WHERE id = $1', [id]);
    if (current.rows.length === 0) return res.status(404).json({ error: 'Not found' });
    const currentPos = current.rows[0].position;

    const adjacent = await pool.query(
      direction === 'up'
        ? 'SELECT id, position FROM todos WHERE position < $1 ORDER BY position DESC LIMIT 1'
        : 'SELECT id, position FROM todos WHERE position > $1 ORDER BY position ASC LIMIT 1',
      [currentPos]
    );

    // Already at the top or bottom—nothing to do
    if (adjacent.rows.length === 0) return res.json({ success: true });

    const adjId = adjacent.rows[0].id;
    const adjPos = adjacent.rows[0].position;

    await pool.query('UPDATE todos SET position = $1 WHERE id = $2', [adjPos, id]);
    await pool.query('UPDATE todos SET position = $1 WHERE id = $2', [currentPos, adjId]);

    res.json({ success: true });
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// Cycle a todo's assignee: '' -> 'M' -> 'D' -> ''
app.post('/todos/:id/assign', async (req, res) => {
  const { id } = req.params;
  try {
    const current = await pool.query('SELECT assignee FROM todos WHERE id = $1', [id]);
    if (current.rows.length === 0) return res.status(404).json({ error: 'Not found' });

    const now = current.rows[0].assignee || '';
    const next = now === '' ? 'M' : now === 'M' ? 'D' : '';

    const result = await pool.query(
      'UPDATE todos SET assignee = $1 WHERE id = $2 RETURNING *',
      [next, id]
    );
    res.json(result.rows[0]);
  } catch (err) {
    res.status(500).json({ error: err.message });
  }
});

// Mark a todo complete/incomplete
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