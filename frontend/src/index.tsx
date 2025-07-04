import React, { useEffect, useState } from 'react';
import { createRoot } from 'react-dom/client';
import axios from 'axios';

interface Message {
  id: string; 
  content: string;
}

function App() {
  const [feed, setFeed] = useState<Message[]>([]);
  const [content, setContent] = useState('');

  useEffect(() => { fetchFeed(); }, []);

  const fetchFeed = async () => {
    try {
      const res = await axios.get('/feed');
      setFeed(res.data);
    } catch (err: unknown) {
      console.error(err);
    }
  };

  const submit = async () => {
    await axios.post('/messages', { content });
    setContent('');
    fetchFeed();
  };

  return (
    <div>
      <h1>Twitter Clone</h1>
      <input value={content} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setContent(e.target.value)} />
      <button onClick={submit}>Post</button>
      {feed.map(m => <div key={m.id}>{m.content}</div>)}
    </div>
  );
}

const root = createRoot(document.getElementById('root')!);
root.render(<App />);
