'use client';

import React, { useState } from 'react';
import { Mail, Lock, LogIn, User, Calendar, UserPlus } from 'lucide-react';

export default function AuthenticationPage() {
  const [role, setRole] = useState<'patient' | 'expert'>('patient');
  const [mode, setMode] = useState<'signin' | 'register'>('signin');

  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4">
      <div className="flex flex-col md:flex-row w-full max-w-[1000px] min-h-[700px] bg-white rounded-3xl shadow-xl overflow-hidden">
        
        <div className="hidden md:block md:w-1/2 relative bg-slate-200">
          <img 
            src="/images/auth-bg.png" 
            alt="Meditation" 
            className="object-cover w-full h-full"
          />
          <div className="absolute top-6 left-6 flex items-center gap-2 text-white font-bold text-xl drop-shadow-md">
            <span className="bg-sky-500 p-1.5 rounded-lg">🧠</span>
            MindCare
          </div>
        </div>

        {/* Right Side - Form */}
        <div className="w-full md:w-1/2 p-8 lg:p-12 flex flex-col justify-center bg-white">
          
          <div className="text-center mb-6">
            <h1 className="text-2xl font-bold text-gray-900 mb-2">
              {mode === 'signin' ? 'Welcome back to MindCare' : 'Create your account'}
            </h1>
            <p className="text-sm text-gray-500">
              {mode === 'signin' 
                ? 'Connect with experts and manage your well-being.' 
                : 'Join our community to start your mental health journey.'}
            </p>
          </div>

          <div className="space-y-5">
            {/* Role Selection */}
            <div>
              <p className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">I am a:</p>
              <div className="flex gap-2">
                {(['patient', 'expert'] as const).map((r) => (
                  <button
                    key={r}
                    onClick={() => setRole(r)}
                    className={`flex-1 py-2 rounded-lg text-sm font-semibold capitalize transition-all ${
                      role === r 
                        ? 'bg-sky-400 text-white shadow-md' 
                        : 'bg-gray-50 text-gray-500 border border-gray-100 hover:bg-gray-100'
                    }`}
                  >
                    {r}
                  </button>
                ))}
              </div>
            </div>

            {/* Mode Switcher */}
            <div className="flex bg-gray-100 p-1 rounded-xl">
              <button
                onClick={() => setMode('signin')}
                className={`flex-1 py-2 rounded-lg text-sm font-bold transition-all ${
                  mode === 'signin' ? 'bg-white text-sky-500 shadow-sm' : 'text-gray-500'
                }`}
              >
                Sign In
              </button>
              <button
                onClick={() => setMode('register')}
                className={`flex-1 py-2 rounded-lg text-sm font-bold transition-all ${
                  mode === 'register' ? 'bg-white text-sky-500 shadow-sm' : 'text-gray-500'
                }`}
              >
                Register
              </button>
            </div>

            {/* Input Fields */}
            <div className="space-y-3">
              {/* Hiển thị Full Name chỉ khi ở chế độ Register */}
              {mode === 'register' && (
                <div className="animate-in fade-in slide-in-from-top-2 duration-300">
                  <label className="block text-xs font-bold text-gray-700 mb-1 ml-1">Full Name</label>
                  <div className="relative">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <User className="h-4 w-4 text-gray-400" />
                    </div>
                    <input
                      type="text"
                      className="block w-full pl-10 pr-3 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-sky-100 focus:border-sky-400 outline-none transition-all text-sm"
                      placeholder="John Doe"
                    />
                  </div>
                </div>
              )}

              <div>
                <label className="block text-xs font-bold text-gray-700 mb-1 ml-1">Email Address</label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Mail className="h-4 w-4 text-gray-400" />
                  </div>
                  <input
                    type="email"
                    className="block w-full pl-10 pr-3 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-sky-100 focus:border-sky-400 outline-none transition-all text-sm"
                    placeholder="name@example.com"
                  />
                </div>
              </div>

              {/* Hiển thị Date of Birth chỉ khi ở chế độ Register */}
              {mode === 'register' && (
                <div className="animate-in fade-in slide-in-from-top-2 duration-300">
                  <label className="block text-xs font-bold text-gray-700 mb-1 ml-1">Date of Birth</label>
                  <div className="relative">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                      <Calendar className="h-4 w-4 text-gray-400" />
                    </div>
                    <input
                      type="date"
                      className="block w-full pl-10 pr-3 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-sky-100 focus:border-sky-400 outline-none transition-all text-sm text-gray-600"
                    />
                  </div>
                </div>
              )}

              <div>
                <label className="block text-xs font-bold text-gray-700 mb-1 ml-1">Password</label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Lock className="h-4 w-4 text-gray-400" />
                  </div>
                  <input
                    type="password"
                    className="block w-full pl-10 pr-3 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-sky-100 focus:border-sky-400 outline-none transition-all text-sm"
                    placeholder="••••••••"
                  />
                </div>
              </div>
            </div>

            {/* Submit Button */}
            <button className="w-full flex items-center justify-center gap-2 bg-sky-400 hover:bg-sky-500 text-white font-bold py-3 rounded-xl transition-all shadow-lg shadow-sky-100 mt-2">
              {mode === 'signin' ? (
                <>
                  <LogIn className="h-5 w-5" /> Sign In
                </>
              ) : (
                <>
                  <UserPlus className="h-5 w-5" /> Create Account
                </>
              )}
            </button>

            {/* Social Divider */}
            <div className="relative flex items-center py-2">
              <div className="flex-grow border-t border-gray-100"></div>
              <span className="flex-shrink-0 mx-4 text-gray-400 text-[10px] font-bold uppercase tracking-widest">Or continue with</span>
              <div className="flex-grow border-t border-gray-100"></div>
            </div>

            {/* Social Buttons */}
            <div className="grid grid-cols-2 gap-3">
              <button className="flex items-center justify-center gap-2 bg-white border border-gray-100 text-gray-600 font-semibold py-2 rounded-xl hover:bg-gray-50 transition-all text-sm">
                <img src="https://www.svgrepo.com/show/475656/google-color.svg" className="w-4 h-4" alt="Google" />
                Google
              </button>
              <button className="flex items-center justify-center gap-2 bg-white border border-gray-100 text-gray-600 font-semibold py-2 rounded-xl hover:bg-gray-50 transition-all text-sm">
                <img src="https://www.svgrepo.com/show/475647/facebook-color.svg" className="w-4 h-4" alt="Facebook" />
                Facebook
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}