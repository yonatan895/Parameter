import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom';
import axios from 'axios';

function App() {
  const [feed, setFeed] = useState<any[]>([]);
  const [content, setContent] = useState('');

  useEffect(() => {
    fetchFeed();
  }, []);

  const fetchFeed = async () => {
    try {
      const res = await axios.get('/feed');
      setFeed(res.data);
    } catch (err) {
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
      <input value={content} onChange={(e) => setContent(e.target.value)} />
      <button onClick={submit}>Post</button>
      {feed.map((m) => (
        <div key={m.id}>{m.content}</div>
      ))}
    </div>
  );
}

ReactDOM.render(<App />, document.getElementById('root'));
