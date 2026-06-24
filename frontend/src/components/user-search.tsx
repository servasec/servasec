import { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import axios from "@/lib/api";

interface SearchUser {
  id: number;
  username: string;
}

interface UserSearchProps {
  value: string;
  onSelect: (userId: string, username: string) => void;
  onClear: () => void;
  placeholder?: string;
  existingIds?: number[];
}

export function UserSearch({ value, onSelect, onClear, placeholder = "Search users (min 2 characters)...", existingIds }: UserSearchProps) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchUser[]>([]);
  const [searching, setSearching] = useState(false);
  const [selectedLabel, setSelectedLabel] = useState("");

  useEffect(() => {
    if (query.length < 2) {
      setResults([]);
      return;
    }
    const timer = setTimeout(() => {
      setSearching(true);
      axios.get(`/api/users/search?q=${encodeURIComponent(query)}`)
        .then((res) => {
          let data = res.data || [];
          if (existingIds) {
            data = data.filter((u: SearchUser) => !existingIds.includes(u.id));
          }
          setResults(data);
        })
        .catch(() => setResults([]))
        .finally(() => setSearching(false));
    }, 300);
    return () => clearTimeout(timer);
  }, [query, existingIds]);

  const handleSelect = (u: SearchUser) => {
    setSelectedLabel(u.username);
    setQuery(u.username);
    setResults([]);
    onSelect(String(u.id), u.username);
  };

  const handleClear = () => {
    setSelectedLabel("");
    setQuery("");
    setResults([]);
    onClear();
  };

  if (value) {
    return (
      <div>
        <p className="text-sm text-muted-foreground">
          Selected: <span className="font-medium text-foreground">{selectedLabel || query}</span>
          {" "}<button type="button" className="text-xs text-primary hover:underline" onClick={handleClear}>Change</button>
        </p>
      </div>
    );
  }

  return (
    <div>
      <Input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder={placeholder}
        autoComplete="off"
        data-1p-ignore
      />
      {searching && (
        <p className="text-xs text-muted-foreground mt-1">Searching...</p>
      )}
      {results.length > 0 && (
        <div className="border rounded-md divide-y max-h-40 overflow-y-auto mt-1">
          {results.map((u) => (
            <button
              key={u.id}
              type="button"
              className="w-full text-left px-3 py-2 text-sm hover:bg-accent transition-colors"
              onClick={() => handleSelect(u)}
            >
              {u.username} <span className="text-muted-foreground">(id: {u.id})</span>
            </button>
          ))}
        </div>
      )}
      {query.length >= 2 && !searching && results.length === 0 && (
        <p className="text-xs text-muted-foreground mt-1">No users found</p>
      )}
    </div>
  );
}
