import React from 'react';
import Navbar from '../components/Navbar';
import HeroSection from '../components/HeroSection';
import FeaturesSection from '../components/FeaturesSection';
import TechStackSection from '../components/TechStackSection';
import QuickStartSection from '../components/QuickStartSection';
import Footer from '../components/Footer';

const Home: React.FC = () => {
  return (
    <div className="min-h-screen bg-slate-900">
      <Navbar />
      <main>
        <section id="home">
          <HeroSection />
        </section>
        <section id="features">
          <FeaturesSection />
        </section>
        <section id="tech">
          <TechStackSection />
        </section>
        <section id="quickstart">
          <QuickStartSection />
        </section>
      </main>
      <Footer />
    </div>
  );
};

export default Home;