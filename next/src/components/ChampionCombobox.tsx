'use client';
import { Combobox, ComboboxInput, ComboboxOption, ComboboxOptions } from '@headlessui/react'
import { Champion } from '@/lib/types/champions';
import { useState } from 'react';
interface Props {
  champions: Champion[]
  onChange: (value: Champion) => void;
  value: Champion | null;
}

export default function ChampionCombobox({champions, onChange, value}: Props) {
  const [query, setQuery] = useState('');

  const filteredChampions = query === ''
    ? champions
    : champions.filter((champion) =>
        champion.name
          .toLowerCase()
          .includes(query.toLowerCase())
      )
  
  return (
    <div className="mt-2">
      <Combobox value={value} onChange={onChange}>
        <ComboboxInput 
          className="border border-gray-300 rounded-md p-2 text-black" 
          displayValue={(champion: Champion | null) => champion?.name ?? ''}
          onChange={(event) => setQuery(event.target.value)}
        />
        <ComboboxOptions>
          {filteredChampions.map(champion => (
            <ComboboxOption key={champion.api_id} value={champion}>
              {champion.name}
            </ComboboxOption>
          ))}
        </ComboboxOptions>
      </Combobox>
    </div>
  )
}