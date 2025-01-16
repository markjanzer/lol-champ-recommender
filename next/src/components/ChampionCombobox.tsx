'use client';
import { Combobox, ComboboxInput, ComboboxOption, ComboboxOptions } from '@headlessui/react'
import { Champion } from '@/lib/types/champions';
interface Props {
  champions: Champion[]
  onChange: (value: Champion) => void;
  value: Champion | null;
}

export default function ChampionCombobox({champions, onChange, value}: Props) {
  return (
    <div className="mt-2">
      <Combobox value={value} onChange={onChange}>
        <ComboboxInput 
          className="border border-gray-300 rounded-md p-2 text-black" 
          displayValue={(champion: Champion | null) => champion?.name ?? ''}
        />
        <ComboboxOptions>
          {champions.map(champion => (
            <ComboboxOption key={champion.api_id} value={champion}>
              {champion.name}
            </ComboboxOption>
          ))}
        </ComboboxOptions>
      </Combobox>
    </div>
  )
}