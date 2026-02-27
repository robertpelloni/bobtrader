import React, { createContext, useContext, useState, useEffect } from 'react';
import { ethers } from 'ethers';

interface WalletContextType {
    address: string | null;
    balance: string | null;
    connect: () => Promise<void>;
    disconnect: () => void;
    isConnecting: boolean;
}

const WalletContext = createContext<WalletContextType | undefined>(undefined);

export const WalletProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [address, setAddress] = useState<string | null>(null);
    const [balance, setBalance] = useState<string | null>(null);
    const [isConnecting, setIsConnecting] = useState(false);

    const connect = async () => {
        if (!window.ethereum) {
            alert("Please install MetaMask!");
            return;
        }
        setIsConnecting(true);
        try {
            const provider = new ethers.BrowserProvider(window.ethereum);
            const signer = await provider.getSigner();
            const addr = await signer.getAddress();
            setAddress(addr);

            const bal = await provider.getBalance(addr);
            setBalance(ethers.formatEther(bal));
        } catch (e) {
            console.error(e);
        } finally {
            setIsConnecting(false);
        }
    };

    const disconnect = () => {
        setAddress(null);
        setBalance(null);
    };

    return (
        <WalletContext.Provider value={{ address, balance, connect, disconnect, isConnecting }}>
            {children}
        </WalletContext.Provider>
    );
};

export const useWallet = () => {
    const context = useContext(WalletContext);
    if (!context) throw new Error("useWallet must be used within a WalletProvider");
    return context;
};
