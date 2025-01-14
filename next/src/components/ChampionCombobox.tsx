'use client';

interface Props {
  champions: {
    name: string;
    api_id: number;
  }[];
  onChange: (value: number) => void;
  value: number;
}

export default function ChampionCombobox({champions, onChange, value}: Props) {
  return (
    <div className="mt-2">
      <input
        className="border border-gray-300 rounded-md p-2 text-black"
        type="number"
        min={1}
        max={10000}
        value={value || ''}
        onChange={(e) => onChange(parseInt(e.target.value) || 0)}
      />
      <span className="text-sm ml-2 font-bold">{champions.find(champion => champion.api_id === value)?.name}</span>
    </div>
  )
}