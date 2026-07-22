# Todo App Setup Guide
## Complete Walkthrough for Mac

---

## Step 1: Create a GitHub Repository

### What this step does:
This step creates a remote backup of your project code on GitHub. GitHub is a cloud-based platform where you can store your code, collaborate with others, and easily deploy your app. Creating a repository is like setting up a folder in the cloud for your project.

### Instructions:
1. Go to github.com
2. Click **New repository**
3. Name it `todo-app`
4. Check "Add a README file"
5. Click **Create repository**
6. Open Terminal and clone it to your computer:
   ```bash
   git clone https://github.com/YOUR_USERNAME/todo-app.git
   cd todo-app
   ```

---

## Step 2: Install Node.js and PostgreSQL (Mac Only)

### What this step does:
Node.js is a runtime that lets you run JavaScript outside of a web browser—it's what powers your backend server. PostgreSQL is a database management system that stores all your data. Homebrew is Mac's package manager that makes installing software easy.

### Install Homebrew (if you don't have it):
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### Install Node.js:
```bash
brew install node
```
Verify: `node --version`

### Install PostgreSQL:
```bash
brew install postgresql@15
```

### Start PostgreSQL:
```bash
brew services start postgresql@15
```

### Verify it works:
```bash
psql --version
```

---

## Step 3: Set Up Node.js Backend Locally

### What this step does:
This step creates the 'engine' of your application. You're installing packages that your backend needs (Express for creating a web server, pg for connecting to PostgreSQL, cors for cross-origin requests, and dotenv for managing secrets). Then you create the .env file to store sensitive configuration and the server.js file which contains the actual backend logic.

### Part 1: Initialize your Node project
```bash
npm init -y
npm install express pg cors dotenv
```

### Part 2: Create .env file
Create a file named `.env` in your `todo-app` folder with this content:
```
DATABASE_URL=postgresql://localhost/todo_app
PORT=3000
```

### Part 3: Create server.js
Create a file named `server.js` in your `todo-app` folder with this content:
```javascript
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
```

---

## Step 4: Set Up the PostgreSQL Database

### What this step does:
This step creates the actual database and table where your todo data will be stored. You're creating a table with columns for id, title, completed status, and timestamp. This is like creating a structured spreadsheet in the cloud that your app can read and write to.

### Open PostgreSQL terminal:
```bash
psql postgres
```

### Create the database and table:
```sql
CREATE DATABASE todo_app;
\c todo_app
CREATE TABLE todos (
  id SERIAL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  completed BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT NOW()
);
```

### Exit PostgreSQL:
```sql
\q
```

---

## Step 5: Create a Simple Frontend

### What this step does:
This step creates the user interface—what users see and interact with in their browser. You're creating an HTML file with styling (CSS) and interactivity (JavaScript). The JavaScript code makes requests to your backend server when users add, delete, or view todos.

### Create index.html:
Create a file named `index.html` in your `todo-app` folder with this content:

```html
<!DOCTYPE html>
<html>
<head>
  <title>Todo App</title>
  <style>
    body { font-family: Arial; max-width: 600px; margin: 50px auto; }
    input { padding: 8px; width: 300px; }
    button { padding: 8px 15px; }
    ul { list-style: none; padding: 0; }
    li { padding: 10px; border: 1px solid #ddd; margin: 5px 0; }
  </style>
</head>
<body>
  <h1>Todo App</h1>
  <input type="text" id="todoInput" placeholder="Add a todo...">
  <button onclick="addTodo()">Add</button>
  <ul id="todoList"></ul>

  <script>
    const API = 'http://localhost:3000';

    async function loadTodos() {
      const res = await fetch(`${API}/todos`);
      const todos = await res.json();
      const list = document.getElementById('todoList');
      list.innerHTML = todos.map(t => `
        <li>
          ${t.title}
          <button onclick="deleteTodo(${t.id})">Delete</button>
        </li>
      `).join('');
    }

    async function addTodo() {
      const input = document.getElementById('todoInput');
      const title = input.value.trim();
      if (!title) return;
      
      await fetch(`${API}/todos`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title })
      });
      input.value = '';
      loadTodos();
    }

    async function deleteTodo(id) {
      await fetch(`${API}/todos/${id}`, { method: 'DELETE' });
      loadTodos();
    }

    loadTodos();
  </script>
</body>
</html>
```

---

## Step 6: Test Locally

### What this step does:
This step starts your backend server and tests that everything works together. You'll run your server locally on your Mac, open it in a browser, and make sure you can add, view, and delete todos without errors.

### Start your server:
```bash
node server.js
```
You should see: `Server running on port 3000`

### Open your browser:
```
http://localhost:3000
```

### Test it:
Try adding a todo, deleting it—it should work! Stop the server with `Ctrl+C` when done.

---

## Step 7: Push to GitHub

### What this step does:
This step uploads all your code to GitHub. You're creating a checkpoint of your work and preparing it to be deployed. 'Push' means uploading your local changes to the remote repository.

```bash
git add .
git commit -m "Initial commit"
git push origin main
```

---

## Step 8: Deploy to Render

### What this step does:
This step sets up your web service on Render. You're telling Render to watch your GitHub repository and automatically run your Node.js server whenever you push new code. This makes your app live on the internet.

1. Go to render.com
2. Sign up with GitHub
3. Click **New +** → **Web Service**
4. Connect your GitHub repo (`todo-app`)
5. Fill in:
   - **Name:** `todo-app`
   - **Runtime:** Node
   - **Build command:** `npm install`
   - **Start command:** `node server.js`
6. Click **Advanced**
7. Don't add DATABASE_URL yet—we'll do that after creating the database

---

## Step 9: Create PostgreSQL Database on Render

### What this step does:
This step creates a managed PostgreSQL database on Render's servers. Instead of you managing a database on your own computer, Render manages it for you. This database is separate from your local one and will be used by your live app.

1. In Render dashboard, click **New +** → **PostgreSQL**
2. Name it `todo-app-db`
3. Select the free tier
4. Click **Create Database**
5. Wait for it to finish (1-2 minutes)
6. Click on the database
7. Copy the connection string (under **Connections** → **Internal Database URL**)

---

## Step 10: Connect Database to Web Service

### What this step does:
This step tells your backend server how to connect to the PostgreSQL database on Render. You're adding an environment variable (DATABASE_URL) so your server knows where to find the database.

1. Go back to your Web Service (click on `todo-app`)
2. Click **Environment**
3. Add environment variable:
   - **Key:** `DATABASE_URL`
   - **Value:** (paste the connection string you just copied from your database)
4. Click **Save Changes**

### Important: Verify the Deployment

After saving, Render will start a redeploy. **Do not test your app until the deployment is complete.**

1. Click the **"Deploys"** tab on your web service
2. Wait for the deployment to finish (you'll see **"Live"** status when done)
3. This usually takes 1-2 minutes

If you test the app before the environment variable is deployed, you'll get 500 errors because your backend can't connect to the database. Make sure the deployment is **"Live"** before testing.

---

## Step 11: Set Up the Database Schema

### What this step does:
This step creates the todos table in your Render PostgreSQL database. It's the same table structure you created locally, but now it's in the cloud where your live app can access it.

### Part 1: Get the External Database URL

1. Go back to your PostgreSQL instance in Render
2. Click **Connect**
3. In the popup that appears, look for the **"External Database URL"** section
4. Copy the entire connection string (it starts with `postgresql://`)

### Part 2: Connect to your database from Terminal

1. Open Terminal on your Mac
2. Paste this command (replace the URL with your copied connection string):
   ```bash
   psql "postgresql://your_copied_url_here"
   ```
   Example:
   ```bash
   psql "postgresql://user:password@host.render.com:5432/dbname"
   ```
3. You should now be connected to your cloud database (you'll see a prompt like `dbname=>`)

### Part 3: Create the table

1. In the psql terminal, paste this SQL command:
   ```sql
   CREATE TABLE todos (
     id SERIAL PRIMARY KEY,
     title VARCHAR(255) NOT NULL,
     completed BOOLEAN DEFAULT FALSE,
     created_at TIMESTAMP DEFAULT NOW()
   );
   ```
2. Press Enter—the table is created!
3. Type `\q` to exit psql

---

## Step 12: Update Frontend for Production

### What this step does:
This step updates your frontend to point to your live app instead of your local computer. When you deployed to Render, it gave you a public URL. Your browser needs to know to send requests to that URL instead of localhost:3000.

1. Edit `index.html` and change this line:
   ```javascript
   const API = 'https://todo-app-abc123.onrender.com'; // Your actual Render URL
   ```
   (Replace with the URL Render gives you in the dashboard)

2. Push to GitHub:
   ```bash
   git add .
   git commit -m "Update API URL for production"
   git push origin main
   ```

Render auto-deploys—your changes are live!

---

## Your App is Now Live! 🎉

Visit your Render URL in a browser and test it. It should work exactly like it did locally.
