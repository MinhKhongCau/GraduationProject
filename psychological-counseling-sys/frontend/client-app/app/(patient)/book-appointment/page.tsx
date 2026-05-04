'use client';

import React, { useState } from 'react';
import { 
  CloudRain, 
  Ghost, 
  HeartCrack, 
  Users, 
  Leaf, 
  BedDouble, 
  Smile, 
  BrainCircuit, 
  Briefcase, 
  MessageSquareQuote,
  Check
} from 'lucide-react';

// Danh sách các triệu chứng/chủ đề
const topics = [
  { id: 'stress', label: 'Stress Management', icon: CloudRain },
  { id: 'anxiety', label: 'Anxiety & Panic', icon: Ghost },
  { id: 'depression', label: 'Depression', icon: HeartCrack },
  { id: 'relationships', label: 'Relationship Issues', icon: Users },
  { id: 'grief', label: 'Grief & Loss', icon: Leaf },
  { id: 'sleep', label: 'Sleep Problems', icon: BedDouble },
  { id: 'self-esteem', label: 'Self-Esteem', icon: Smile },
  { id: 'trauma', label: 'Trauma & PTSD', icon: BrainCircuit },
  { id: 'work-life', label: 'Work-Life Balance', icon: Briefcase },
  { id: 'communication', label: 'Communication Skills', icon: MessageSquareQuote },
];

export default function BookAppointmentPage() {
  // State lưu trữ các ID của chủ đề được chọn
  const [selectedTopics, setSelectedTopics] = useState<string[]>([]);

  // Hàm xử lý khi click vào 1 ô
  const toggleTopic = (id: string) => {
    setSelectedTopics((prev) => 
      prev.includes(id) 
        ? prev.filter((topicId) => topicId !== id) // Nếu đã chọn thì bỏ chọn
        : [...prev, id] // Nếu chưa chọn thì thêm vào
    );
  };

  return (
    <div className="max-w-5xl mx-auto h-full flex flex-col">
      {/* Header Section */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-slate-900 mb-3">What brings you here today?</h1>
        <p className="text-slate-500 text-sm max-w-2xl">
            Please select the topics or symptoms you&apos;d like to discuss with an expert. <br/>
            You can choose multiple.
        </p>
      </div>

      {/* Grid of Topics */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-8">
        {topics.map((topic) => {
          const isSelected = selectedTopics.includes(topic.id);
          const Icon = topic.icon;

          return (
            <button
              key={topic.id}
              onClick={() => toggleTopic(topic.id)}
              className={`relative flex flex-col items-center justify-center p-6 rounded-2xl border transition-all duration-200 group h-36 ${
                isSelected 
                  ? 'border-sky-500 bg-sky-50/50 shadow-md shadow-sky-100/50' 
                  : 'border-slate-200 bg-white hover:border-sky-300 hover:shadow-sm'
              }`}
            >
              {/* Checkmark icon khi được chọn */}
              {isSelected && (
                <div className="absolute top-3 right-3 text-sky-500 animate-in zoom-in duration-200">
                  <Check className="w-4 h-4 stroke-[3]" />
                </div>
              )}
              
              <Icon 
                className={`w-8 h-8 mb-4 transition-colors ${
                  isSelected ? 'text-sky-500' : 'text-slate-600 group-hover:text-sky-400'
                }`} 
                strokeWidth={1.5}
              />
              <span className={`text-sm font-semibold text-center ${
                isSelected ? 'text-sky-700' : 'text-slate-700'
              }`}>
                {topic.label}
              </span>
            </button>
          );
        })}
      </div>

      {/* Nút Next ở góc phải dưới */}
      <div className="mt-auto flex justify-end border-t border-slate-200 pt-6">
        <button 
          disabled={selectedTopics.length === 0}
          className={`flex items-center gap-2 px-8 py-2.5 rounded-lg font-semibold transition-all ${
            selectedTopics.length > 0
              ? 'bg-sky-500 hover:bg-sky-600 text-white shadow-md'
              : 'bg-slate-100 text-slate-400 cursor-not-allowed'
          }`}
        >
          Next
          <Check className="w-4 h-4" />
        </button>
      </div>
    </div>
  );
}