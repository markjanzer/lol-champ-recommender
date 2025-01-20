'use client';
import { Combobox, ComboboxInput, ComboboxOption, ComboboxOptions } from '@headlessui/react'
import { Champion } from '@/lib/types/champions';
import { useState } from 'react';
import Fuse from 'fuse.js';
interface Props {
  champions: Champion[]
  onChange: (value: Champion) => void;
  value: Champion | null;
}

export default function ChampionCombobox({champions, onChange, value}: Props) {
  const [query, setQuery] = useState('');
  
  const fuse = new Fuse(champions, {
    keys: ['name'],
    threshold: 0.3
  }); 
  
  const fuseResults = fuse.search(query)
  const filteredChampions = fuseResults.map(result => result.item)

    
  return (
    <div className="mt-2">
      <Combobox value={value} onChange={onChange}>
        <ComboboxInput 
          className="border border-gray-300 rounded-md p-2 text-black" 
          displayValue={(champion: Champion | null) => champion?.name ?? ''}
          onChange={(event) => setQuery(event.target.value)}
        />
        <ComboboxOptions
          anchor="bottom"
          className="border empty:invisible w-[var(--input-width)]"
        >
          {filteredChampions.map(champion => (
            <ComboboxOption 
              key={champion.api_id} 
              value={champion}
              className=" p-2 text-black data-[focus]:bg-gray-200 bg-white"
            >
              {champion.name}
            </ComboboxOption>
          ))}
        </ComboboxOptions>
      </Combobox>
    </div>
  )
}