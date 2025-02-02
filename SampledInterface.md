# Sampled interface

## Purpose

The purpose of the **Sampled** interface is to create a function, GetSampled(song, artist string) (SpotifyURI, error), that can be implemented regardless of the source (ie, Genius, OpenAI, etc.). 

## Sampled Manager

The **Sampled Manager** will take a list of sources that implement the *Sampled* interface.

The order of the Sampled Manager will determine the priority of the source we want to get the sample from. 


