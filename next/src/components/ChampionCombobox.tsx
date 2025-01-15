'use client';
import { Combobox, ComboboxInput, ComboboxOption, ComboboxOptions } from '@headlessui/react'
import { useState } from 'react';

interface Props {
  champions: {
    name: string;
    api_id: number;
  }[];
  onChange: (value: number) => void;
  value: number;
}

export default function ChampionCombobox({champions, onChange, value}: Props) {
  const [selectedChampionId, setSelectedChampionId] = useState(value);

  const handleChange = (championId: number) => {
    setSelectedChampionId(championId);
    onChange(championId);
  }
  
  return (
    <div className="mt-2">
      <Combobox value={selectedChampionId} onChange={handleChange}>
        <ComboboxInput 
          className="border border-gray-300 rounded-md p-2 text-black" 
          displayValue={(championId: number) => {
            const champion = champions.find(champion => champion.api_id === championId);
            return champion ? champion.name : '';
          }}
        />
        <ComboboxOptions>
          {champions.map(champion => (
            <ComboboxOption key={champion.api_id} value={champion.api_id}>{champion.name}</ComboboxOption>
          ))}
        </ComboboxOptions>
      </Combobox>
    </div>
  )
}