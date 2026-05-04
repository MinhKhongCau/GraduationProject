'use client';

import React from 'react';
import { Brain, CloudLightning, Activity, Zap } from 'lucide-react';

// Dữ liệu giả lập (Sau này sẽ fetch từ Assessment Service)
const assessments = [
  {
    id: 'phq9',
    title: 'Depression Screening',
    description: 'Assess your mood and identify potential signs of depression with this quick evaluation.',
    questions: 10,
    icon: Brain,
  },
  {
    id: 'gad7',
    title: 'Anxiety Assessment',
    description: 'Evaluate your anxiety levels and common triggers to better understand your stress responses.',
    questions: 8,
    icon: CloudLightning,
  },
  {
    id: 'stress',
    title: 'Stress Level Check',
    description: 'Understand your current stress and explore effective coping mechanisms tailored for you.',
    questions: 12,
    icon: Activity,
  },
  {
    id: 'asrs',
    title: 'ADHD Self-Assessment',
    description: 'Explore symptoms related to Attention-Deficit/Hyperactivity Disorder for initial insights.',
    questions: 15,
    icon: Zap,
  }
];

export default function AssessmentPage() {
  return (
    <div className="max-w-7xl mx-auto h-full flex flex-col items-center py-6">
      
      {/* Tiêu đề & Cấu trúc */}
      <div className="text-center mb-12 max-w-3xl">
        <h1 className="text-5xl font-bold text-[#56b0f9] mb-6">
          Explore Your Inner Landscape
        </h1>
        <p className="text-slate-600 text-lg leading-relaxed px-4">
          Our curated library of psychological tests offers a supportive starting point to
          understand your mental well-being. Select a topic to begin your journey of
          self-discovery.
        </p>
      </div>

      {/* Grid danh sách bài Test */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 w-full">
        {assessments.map((test) => {
          const Icon = test.icon;
          
          return (
            <div 
              key={test.id} 
              className="bg-white border border-slate-100 rounded-3xl p-6 flex flex-col hover:shadow-[0_8px_30px_rgb(0,0,0,0.04)] hover:border-sky-100 transition-all duration-300"
            >
              {/* Icon */}
              <div className="w-14 h-14 bg-sky-50/80 rounded-full flex items-center justify-center mb-6">
                <Icon className="w-6 h-6 text-[#56b0f9]" strokeWidth={2} />
              </div>

              {/* Thông tin */}
              <h2 className="text-xl font-bold text-slate-900 mb-3">{test.title}</h2>
              <p className="text-sm text-slate-600 leading-relaxed mb-6 flex-grow">
                {test.description}
              </p>

              {/* Số lượng câu hỏi */}
              <p className="text-sm font-semibold text-slate-600 mb-5">
                {test.questions} Questions
              </p>

              {/* Nút bấm Start */}
              <button className="w-full bg-[#5fb2f9] hover:bg-[#4a9ee6] text-white font-semibold py-3 rounded-xl transition-colors shadow-sm">
                Start Test
              </button>
            </div>
          );
        })}
      </div>

    </div>
  );
}